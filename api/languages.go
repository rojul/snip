package api

import (
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/go-yaml/yaml"
	"github.com/gorilla/mux"
)

var (
	HTTPErrorLanguageNotFound = HTTPError{Status: http.StatusNotFound, Msg: "Language Not Found"}
)

type languagesObj struct {
	Languages []*Language `json:"languages" yaml:"languages"`
}

func (h *handler) languagesRouter(r *mux.Router) {
	r.HandleFunc("", h.languageListHandler).Methods("GET")
	r.HandleFunc("/{id}", h.languageHandler).Methods("GET")
}

func (h *handler) languageListHandler(w http.ResponseWriter, r *http.Request) {
	ls := h.GetLanguages()
	bls := make([]*Language, len(ls))
	for i, l := range ls {
		bls[i] = &Language{
			ID:        l.ID,
			Name:      l.Name,
			Extension: l.Extension,
		}
	}

	sendJSON(w, &languagesObj{
		Languages: bls,
	})
}

func (h *handler) languageHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	language, err := h.getLanguage(id)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, language)
}

func (h *handler) getLanguage(id string) (*Language, error) {
	for _, language := range h.GetLanguages() {
		if language.ID == id {
			return language, nil
		}
	}
	return nil, HTTPErrorLanguageNotFound
}

func (h *handler) GetLanguages() []*Language {
	return h.languages
}

func loadLanguagesYaml(file string) ([]*Language, error) {
	var obj languagesObj
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	languages := obj.Languages
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].ID < languages[j].ID
	})
	return languages, nil
}
