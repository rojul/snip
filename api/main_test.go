package api

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/client"
	units "github.com/docker/go-units"
)

var testH *handler
var testLanguages = flag.Bool("languages", false, "test all languages")

func TestMain(m *testing.M) {
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

	h.config = &Config{
		RunTimeout:         30 * time.Second,
		Memory:             512 * units.MiB,
		SnippetSizeLimit:   1 * units.MiB,
		DefaultImagePrefix: "snip",
		LanguagesFile:      "languages.yml",
		ReturnSizeLimit:    100 * units.KiB,
	}

	if h.languages, err = loadLanguagesYaml(h.config.LanguagesFile); err != nil {
		return nil, err
	}

	if h.dockerClient, err = client.NewEnvClient(); err != nil {
		return nil, err
	}
	return h, nil
}
