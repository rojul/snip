package runner

import (
	"sync"
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
	res *Result
	t   EventType
	mu  *sync.Mutex
}

func (w *eventWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.res.Append(&Event{
		Type:    w.t,
		Message: string(p),
		Time:    time.Now(),
	})
	return len(p), nil
}
