package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
)

func JSONRecover() {
	if r := recover(); r != nil {
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"error": fmt.Sprintf("panic in runner: %v\n\n%s", r, debug.Stack()),
		})
		os.Exit(1)
	}
}

func Run(r io.Reader) *Result {
	var payload Payload
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return &Result{Error: "Failed to parse input json: " + err.Error()}
	}

	if err := writeFiles(payload.Files); err != nil {
		return &Result{Error: "Failed to write file to disk: " + err.Error()}
	}

	return runCommand(&payload)
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

func runCommand(payload *Payload) (res *Result) {
	res = &Result{}

	cmd := exec.Command("sh", "-c", payload.Command)
	cmd.Stdin = strings.NewReader(payload.Stdin)

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	cmd.Env = os.Environ()
	if len(payload.Files) > 0 {
		cmd.Env = append(cmd.Env, "FILE="+payload.Files[0].Name)
	}
	err := cmd.Run()

	if b.Len() != 0 {
		res.Append(&Event{
			Type:    Stdout,
			Message: b.String(),
		})
	}

	res.ExitCode = getExitCode(err)

	if res.ExitCode == nil || *res.ExitCode < 0 {
		res.Error = err.Error()
	}

	return
}
