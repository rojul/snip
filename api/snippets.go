package api

import (
	"net/http"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
)

func (h *handler) getDatabase() *mgo.Database {
	return h.mgoClient.DB(h.config.MongoDB)
}

func (h *handler) getSnippetCollection() *mgo.Collection {
	return h.getDatabase().C("snippets")
}

func (h *handler) snippetsRouter(r *mux.Router) {
	r.HandleFunc("", h.createSnippetsHandler).Methods("POST")
	r.HandleFunc("/{id}", h.getSnippetsHandler).Methods("GET")
}

func (h *handler) createSnippetsHandler(w http.ResponseWriter, r *http.Request) {
	var snippet Snippet
	if ok := readJSONBody(w, r, h.config.SnippetSizeLimit, &snippet); !ok {
		return
	}

	if err := snippet.getValidationError(); err != nil {
		sendError(w, HTTPError{Status: http.StatusBadRequest, Msg: "Invalid payload: " + err.Error()})
		return
	}

	snippet.ID = bson.NewObjectId()
	snippet.Created = time.Now()
	snippet.Modified = snippet.Created

	if err := h.putSnippet(&snippet); err != nil {
		sendError(w, err)
		return
	}

	sendJSON(w, &snippet)
}

func (h *handler) putSnippet(snippet *Snippet) error {
	if err := h.getSnippetCollection().Insert(*snippet); err != nil {
		return err
	}
	return nil
}

func (h *handler) getSnippetsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseSnippetID(mux.Vars(r)["id"])
	if err != nil {
		sendError(w, HTTPErrorInvalidSnippetID)
		return
	}

	snippet, err := h.getSnippet(id)
	if err != nil {
		sendError(w, err)
		return
	}

	sendJSON(w, snippet)
}

func (h *handler) getSnippet(id bson.ObjectId) (*Snippet, error) {
	var snippet Snippet
	err := h.getSnippetCollection().FindId(id).One(&snippet)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, HTTPErrorSnippetNotFound
		}
		return nil, err
	}

	return &snippet, nil
}
