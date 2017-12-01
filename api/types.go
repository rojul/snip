package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rojul/snip/api/runner"
	"gopkg.in/mgo.v2/bson"
)

var (
	HTTPErrorInvalidSnippetID = HTTPError{Status: http.StatusBadRequest, Msg: "Invalid Snippet ID"}
	HTTPErrorSnippetNotFound  = HTTPError{Status: http.StatusNotFound, Msg: "Snippet Not Found"}
)

type Language struct {
	ID          string                  `json:"id" toml:"id"`
	Name        string                  `json:"name,omitempty" toml:"name"`
	Extension   string                  `json:"extension,omitempty" toml:"extension"`
	Command     string                  `json:"command,omitempty" toml:"command"`
	Image       string                  `json:"image,omitempty" toml:"image"`
	NotRunnable bool                    `json:"notRunnable,omitempty" toml:"notRunnable"`
	Tests       map[string]LanguageTest `json:"tests,omitempty" toml:"tests"`
}

func (l *Language) getTestPayload(name string) runner.Payload {
	t := l.Tests[name]
	p := runner.Payload{
		Stdin:   t["_stdin"],
		Command: t["_command"],
	}
	if main, ok := t["_main"]; ok {
		p.Files = append(p.Files, &runner.File{
			Name:    "main." + l.Extension,
			Content: main,
		})
	}
	var keys []string
	for k := range t {
		if !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		p.Files = append(p.Files, &runner.File{
			Name:    k,
			Content: t[k],
		})
	}
	return p
}

type LanguageTest map[string]string

type Payload struct {
	runner.Payload `bson:",inline"`
	Language       string `json:"language,omitempty" bson:",omitempty"`
}

func (p *Payload) getValidationError() error {
	if len(p.Language) > 64 {
		return errors.New("Language ID too long")
	}
	if len(p.Files) > 10 {
		return errors.New("Too many files")
	}
	if len(p.Files) == 0 {
		return errors.New("At least one file required")
	}
	for i, file := range p.Files {
		if file.Name == "" {
			return errors.New("Filename required for file " + strconv.Itoa(i+1))
		}
	}
	return nil
}

type Snippet struct {
	Payload  `bson:",inline"`
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Created  time.Time     `json:"created"`
	Modified time.Time     `json:"modified"`
	Public   bool          `json:"public,omitempty" bson:",omitempty"`
}

func (s *Snippet) MarshalJSON() ([]byte, error) {
	type Alias Snippet
	return json.Marshal(&struct {
		ID       string `json:"id"`
		Created  int64  `json:"created"`
		Modified int64  `json:"modified"`
		*Alias
	}{
		ID:       s.ID.Hex(),
		Created:  s.Created.Unix(),
		Modified: s.Modified.Unix(),
		Alias:    (*Alias)(s),
	})
}

func parseSnippetID(id string) (bson.ObjectId, error) {
	if !bson.IsObjectIdHex(id) {
		return "", HTTPErrorInvalidSnippetID
	}
	return bson.ObjectIdHex(id), nil
}
