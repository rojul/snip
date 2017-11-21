package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
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

var (
	errOutputTruncated = errors.New("output truncated")
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

	result, err := h.runContainerSync(&payload, language)
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

func (h *handler) runContainer(payload *Payload, language *Language, events chan<- *runner.Event) (*runner.Result, error) {
	defer close(events)
	if language.NotRunnable {
		return &runner.Result{Error: "This language is not runnable"}, nil
	}
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
		User:            "1000:1000",
	}
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		CapDrop:    []string{"ALL"},
		Resources: container.Resources{
			Memory:     h.config.Memory,
			MemorySwap: h.config.Memory,
			NanoCPUs:   h.config.NanoCPUs,
			CPUShares:  h.config.CPUShares,
			PidsLimit:  h.config.PidsLimit,
		},
		LogConfig: container.LogConfig{
			Type: "none",
		},
		Tmpfs: map[string]string{
			"/tmp":       "exec",
			"/home/snip": "exec,uid=1000,gid=1000",
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

	result := &runner.Result{}

	done := make(chan bool)
	lines := make(chan []byte)
	go func() {
		for l := range lines {
			var e runner.Event
			if err := json.Unmarshal(l, &e); err == nil && e.Type != "" {
				events <- &e
			} else if err := json.Unmarshal(l, result); err == nil && !result.IsEmpty() {
			} else {
				result = &runner.Result{Error: "Invalid response: " + string(l)}
				break
			}
		}
		done <- true
	}()

	stderr, err := collectDockerStream(res.Reader, h.config.ReturnSizeLimit, lines)
	<-done
	if stderr != "" {
		if err == errOutputTruncated {
			stderr += "\n[truncated]"
		}
		return &runner.Result{Error: "Container returned an error:\n\n" + stderr}, nil
	}
	if err == errOutputTruncated {
		return &runner.Result{Error: "Output truncated"}, nil
	}
	if err != nil {
		return nil, err
	}
	if ctx.Err() == context.DeadlineExceeded {
		return &runner.Result{Error: "Container timed out"}, nil
	}
	if result.IsEmpty() {
		result = &runner.Result{Error: "No response from container"}
	}

	return result, nil
}

func (h *handler) runContainerSync(payload *Payload, language *Language) (*runner.Result, error) {
	events := make(chan *runner.Event)
	done := make(chan bool)
	var es []*runner.Event
	go func() {
		for e := range events {
			es = append(es, e)
		}
		done <- true
	}()
	r, err := h.runContainer(payload, language, events)
	<-done
	r.Events = es
	return r, err
}

func collectDockerStream(stream io.Reader, limit int64, lines chan<- []byte) (string, error) {
	defer close(lines)
	var n int64
	var stderr bytes.Buffer
	var buf bytes.Buffer
	header := make([]byte, 8)
	for {
		n += 8
		if n > limit {
			return stderr.String(), errOutputTruncated
		}
		if _, err := io.ReadFull(stream, header); err != nil {
			if err == io.EOF {
				return stderr.String(), nil
			}
			return "", err
		}

		var w io.Writer
		if header[0] == 1 {
			w = bufio.NewWriter(&buf)
		} else if header[0] == 2 {
			w = bufio.NewWriter(&stderr)
		} else {
			return "", fmt.Errorf("invalid STREAM_TYPE: %x", header[0])
		}

		frameSize := int64(binary.BigEndian.Uint32(header[4:]))
		n += frameSize
		if n > limit {
			frameSize -= n - limit
		}

		if _, err := io.CopyN(w, stream, frameSize); err != nil {
			return "", err
		}

		s := bufio.NewScanner(bufio.NewReader(&buf))
		for s.Scan() {
			lines <- s.Bytes()
		}
		if err := s.Err(); err != nil {
			return "", err
		}

		if n > limit {
			return stderr.String(), errOutputTruncated
		}
	}
}
