package api

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/docker/docker/client"
)

var testH *handler
var testLanguages = flag.Bool("languages", false, "test all languages")

func TestMain(m *testing.M) {
	if testing.Short() {
		return
	}

	var err error
	testH, err = newTestHandler()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func newTestHandler() (h *handler, err error) {
	h = &handler{}

	h.config = defaultConfig()

	if h.languages, err = loadLanguagesYaml(h.config.LanguagesFile); err != nil {
		return nil, err
	}

	if h.dockerClient, err = client.NewEnvClient(); err != nil {
		return nil, err
	}
	return h, nil
}
