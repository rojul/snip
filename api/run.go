package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gorilla/mux"
	"github.com/rojul/snip/api/runner"
)

func (h *handler) runRouter(r *mux.Router) {
	r.HandleFunc("", h.runHandler).Methods("POST")
}

func (h *handler) runHandler(w http.ResponseWriter, r *http.Request) {
	var payload Payload
	if ok := readJSONBody(w, r, h.config.SnippetSizeLimit, &payload); !ok {
		return
	}

	language, err := h.getLanguage(payload.Language)
	if err != nil {
		sendError(w, err)
		return
	}

	if err := payload.getValidationError(); err != nil {
		sendError(w, HTTPError{Status: http.StatusBadRequest, Msg: "Invalid payload: " + err.Error()})
		return
	}

	result, err := h.runContainer(&payload, language)
	if err != nil {
		sendError(w, err)
		return
	}

	sendJSON(w, result)
}

func (h *handler) removeContainer(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.dockerClient.ContainerRemove(ctx, id, dockerTypes.ContainerRemoveOptions{Force: true}); err != nil {
		if strings.Contains(err.Error(), "is already in progress") ||
			strings.Contains(err.Error(), "No such container") {
			return
		}
		log.Debug("killing container failed: " + err.Error())
	}
}

func (h *handler) runContainer(payload *Payload, language *Language) (*runner.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.RunTimeout)
	defer cancel()

	if payload.Command == "" {
		payload.Command = language.Command
	}

	image := language.Image
	if image == "" {
		image = h.config.DefaultImagePrefix + "/" + language.ID
	}

	containerConfig := &container.Config{
		Image:           image,
		AttachStdin:     true,
		AttachStdout:    true,
		AttachStderr:    true,
		OpenStdin:       true,
		StdinOnce:       true,
		NetworkDisabled: !h.config.NetworkEnabled,
		User:            "snip:snip",
	}
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		CapDrop:    []string{"ALL"},
		Resources: container.Resources{
			Memory:     h.config.Memory,
			MemorySwap: h.config.Memory,
			NanoCPUs:   h.config.NanoCPUs,
			CPUShares:  512,
			PidsLimit:  35,
		},
		LogConfig: container.LogConfig{
			Type: "none",
		},
		Tmpfs: map[string]string{
			"/tmp":       "exec",
			"/home/snip": "exec",
		},
		ReadonlyRootfs: true,
	}

	c, err := h.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"id":       c.ID[:12],
		"language": language.ID,
	}).Debug("container started")

	go func() {
		<-ctx.Done()
		h.removeContainer(c.ID)
	}()

	attachOptions := dockerTypes.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	}

	res, err := h.dockerClient.ContainerAttach(ctx, c.ID, attachOptions)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	err = h.dockerClient.ContainerStart(ctx, c.ID, dockerTypes.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	payloadBytes, err := json.Marshal(payload.Payload)
	if err != nil {
		return nil, err
	}
	res.Conn.Write(payloadBytes)
	res.CloseWrite()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(io.LimitReader(res.Reader, h.config.ReturnSizeLimit)); err != nil {
		return nil, err
	}
	bufStr := buf.Bytes()

	if len(bufStr) >= 8 {
		// remove stream header
		bufStr = bufStr[8:]
	}

	result := &runner.Result{}
	err = json.Unmarshal(bufStr, result)
	if err != nil {
		if h.config.ReturnSizeLimit == int64(buf.Len()) {
			return &runner.Result{Error: "Output too large"}, nil
		}
		if ctx.Err() == context.DeadlineExceeded {
			return &runner.Result{Error: "Container timed out"}, nil
		}
		return &runner.Result{Error: "Invalid response from container"}, nil
	}

	return result, nil
}
