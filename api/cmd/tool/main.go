package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/rojul/snip/api"
)

var languageIDPattern = regexp.MustCompile("^[a-z0-9]+$")

func main() {
	dir := "./languages"
	ids, err := getLanguageIDs(dir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ls := make([]*api.Language, len(ids))
	ok := true
	for i, id := range ids {
		l, err := handleLanguage(dir, id)
		if err != nil {
			ok = false
			fmt.Printf("%s: %v\n", id, err)
			continue
		}
		ls[i] = l
	}
	if !ok {
		os.Exit(1)
	}

	if err := writeLanguagesJSON(ls); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("done")
}

func getLanguageIDs(dirname string) ([]string, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, f := range files {
		if languageIDPattern.MatchString(f.Name()) && f.IsDir() {
			ids = append(ids, f.Name())
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("no languages found")
	}
	return ids, nil
}

func handleLanguage(dirname, id string) (*api.Language, error) {
	b, err := ioutil.ReadFile(path.Join(dirname, id, "config.toml"))
	if err != nil {
		return nil, err
	}
	l := &api.Language{}
	err = toml.Unmarshal(b, l)
	if err != nil {
		return nil, err
	}
	l.ID = id
	return l, nil
}

func writeLanguagesJSON(ls []*api.Language) error {
	b, err := json.MarshalIndent(&struct {
		Languages []*api.Language `json:"languages"`
	}{ls}, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("./api/languages.json", b, 0664)
}
