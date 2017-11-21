package runner

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func writeJSON(w io.Writer, v interface{}) {
	json.NewEncoder(w).Encode(v)
}

func Run(r io.Reader, w io.Writer) {
	var payload Payload
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		writeJSON(w, &Result{Error: "Failed to parse input json: " + err.Error()})
		return
	}

	if err := writeFiles(payload.Files); err != nil {
		writeJSON(w, &Result{Error: "Failed to write file to disk: " + err.Error()})
		return
	}

	runCommand(w, &payload)
}

func writeFiles(files []*File) error {
	for _, file := range files {
		if err := writeFile(file); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(file *File) error {
	if err := os.MkdirAll(filepath.Dir(file.Name), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(file.Name, []byte(file.Content), 0755); err != nil {
		return err
	}

	return nil
}

func getExitCode(err error) *int {
	if err == nil {
		c := 0
		return &c
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		ws := exitError.Sys().(syscall.WaitStatus)
		c := ws.ExitStatus()
		return &c
	}
	return nil
}

func runCommand(w io.Writer, payload *Payload) {
	cmd := exec.Command("sh", "-c", payload.Command)
	cmd.Stdin = strings.NewReader(payload.Stdin)

	ew := &eventWriter{w}
	cmd.Stdout = ew
	cmd.Stderr = ew
	cmd.Env = os.Environ()
	if len(payload.Files) > 0 {
		cmd.Env = append(cmd.Env, "FILE="+payload.Files[0].Name)
	}
	err := cmd.Run()

	res := &Result{}
	res.ExitCode = getExitCode(err)

	if res.ExitCode == nil || *res.ExitCode < 0 {
		res.Error = err.Error()
	}

	writeJSON(w, res)
}
