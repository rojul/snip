package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rojul/snip/api"
)

func main() {
	log.SetLevel(log.DebugLevel)
	h, err := api.NewDefaultServer()
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()
	log.Fatal(h.Serve())
}
