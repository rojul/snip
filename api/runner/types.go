package runner

import (
	"io"
	"time"
)

type EventType string

const (
	Stdout EventType = "stdout"
	Stderr EventType = "stderr"
)

type Event struct {
	Type    EventType `json:"type"`
	Message string    `json:"message"`
	Time    time.Time `json:"-"`
}

type Result struct {
	Events   []*Event `json:"events,omitempty"`
	Error    string   `json:"error,omitempty"`
	ExitCode *int     `json:"exitCode,omitempty"`
}

func (res *Result) Append(e *Event) {
	res.Events = append(res.Events, e)
}

func (res *Result) IsEmpty() bool {
	return res.Error == "" && res.ExitCode == nil
}

type Payload struct {
	Files   []*File `json:"files"`
	Stdin   string  `json:"stdin,omitempty" bson:",omitempty"`
	Command string  `json:"command,omitempty" bson:",omitempty"`
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type eventWriter struct {
	w io.Writer
}

func (ew *eventWriter) Write(b []byte) (n int, err error) {
	writeJSON(ew.w, &Event{
		Type:    Stdout,
		Message: string(b),
	})

	return len(b), nil
}
