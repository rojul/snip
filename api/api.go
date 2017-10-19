package api

import (
	"net/http"
	"runtime/debug"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
)

const version = "0.1.0"

type handler struct {
	config       *Config
	languages    []*Language
	dockerClient *client.Client
	mgoClient    *mgo.Session
}

func (h *handler) homeHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, &Fields{
		"description": "Snip Api",
		"version":     version,
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, HTTPError{Status: http.StatusNotFound, Msg: "Resousnip Not Found"})
}

func addSubrouter(r *mux.Router, tpl string, f func(*mux.Router)) {
	f(r.PathPrefix(tpl).Subrouter())
}

type logrusRecoveryHandlerLogger struct{}

func (l logrusRecoveryHandlerLogger) Println(v ...interface{}) {
	log.WithField("stack", string(debug.Stack())).Error(v)
}

func (h *handler) getAPIHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", h.homeHandler).Methods("GET")
	addSubrouter(r, "/run", h.runRouter)
	addSubrouter(r, "/languages", h.languagesRouter)
	addSubrouter(r, "/snippets", h.snippetsRouter)
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	hh := handlers.CompressHandler(r)
	if h.config.CorsEnabled {
		hh = handlers.CORS()(hh)
	}

	rl := handlers.RecoveryLogger(&logrusRecoveryHandlerLogger{})
	hh = handlers.RecoveryHandler(rl)(hh)
	return hh
}

func (h *handler) Serve() error {
	timeout := 10 * time.Second
	srv := &http.Server{
		Handler:      h.getAPIHandler(),
		Addr:         h.config.HTTPAddr,
		ReadTimeout:  timeout,
		WriteTimeout: timeout + h.config.RunTimeout,
	}
	return srv.ListenAndServe()
}

func NewDefaultServer() (h *handler, err error) {
	h = &handler{}

	if h.config, err = configFromEnv(); err != nil {
		return nil, err
	}

	if h.config.JSONLogging {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	}
	log.Info("server starting")

	if h.languages, err = loadLanguagesYaml(h.config.LanguagesFile); err != nil {
		return nil, err
	}

	if h.dockerClient, err = client.NewEnvClient(); err != nil {
		return nil, err
	}

	if h.mgoClient, err = mgo.Dial(h.config.MongoURL); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *handler) Close() {
	h.mgoClient.Close()
}
