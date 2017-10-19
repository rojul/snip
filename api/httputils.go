package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type Fields map[string]interface{}

type HTTPError struct {
	Status int    `json:"-"`
	Msg    string `json:"error"`
	Reason string `json:"reason,omitempty"`
}

func (e HTTPError) Error() string {
	return e.Msg
}

func sendError(w http.ResponseWriter, err error) {
	httpErr, ok := err.(HTTPError)
	if !ok {
		log.Error(err.Error())
	} else {
		log.Debug(err.Error())
	}

	if httpErr.Status == 0 {
		httpErr.Status = http.StatusInternalServerError
	}
	if httpErr.Msg == "" {
		httpErr.Msg = http.StatusText(httpErr.Status)
	}

	sendJSONWithStatus(w, httpErr.Status, httpErr)
}

func sendJSON(w http.ResponseWriter, v interface{}) {
	sendJSONWithStatus(w, http.StatusOK, v)
}

func sendJSONWithStatus(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func readJSONBody(w http.ResponseWriter, r *http.Request, n int64, v interface{}) (ok bool) {
	if n > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, n)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			sendError(w, HTTPError{Status: http.StatusRequestEntityTooLarge})
			return
		}
		sendError(w, err)
		return
	}
	if err := r.Body.Close(); err != nil {
		sendError(w, err)
		return
	}
	if err := json.Unmarshal(body, v); err != nil {
		sendError(w, HTTPError{Status: http.StatusBadRequest, Msg: "Invalid JSON", Reason: err.Error()})
		return
	}
	return true
}
