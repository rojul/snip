package api

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/rojul/snip/api/runner"
)

func TestLanguageTests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var ls []*Language
	if !*testLanguages {
		t.Log("pass -languages to test all languages")
		ls = []*Language{mustGetAsh()}
	} else {
		ls = testH.GetLanguages()
	}

	for _, l := range ls {
		keys := make([]string, 0, len(l.Tests))
		for k := range l.Tests {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			t.Run(l.ID+"/"+k, func(t *testing.T) { testLanguageTest(t, l, k) })
		}
	}
}

func testLanguageTest(t *testing.T, l *Language, name string) {
	if l.NotRunnable {
		t.Skip("not runnable")
	}
	p := &Payload{
		Language: l.ID,
		Payload:  l.getTestPayload(name),
	}
	stdout := l.Tests[name]["_stdout"]
	if stdout == "" {
		stdout = "Hello World\n"
	}
	r, err := testH.runContainerSync(p, l)
	if err != nil {
		t.Fatal(err)
	}
	rJSON := mustToJSON(r)
	if !compareResult(r, stdout) {
		t.Fatalf("\nexpected: %#v\nactual:   %s", stdout, rJSON)
	}
	if *verbose {
		t.Logf("ok: %s", rJSON)
	}
}

func mustGetAsh() *Language {
	l, err := testH.getLanguage("ash")
	if err != nil {
		panic(err)
	}
	return l
}

func mustToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func compareResult(r *runner.Result, stdout string) bool {
	rStdout := ""
	for _, e := range r.Events {
		switch e.Type {
		case runner.Stdout:
			rStdout += e.Message
		default:
			return false
		}
	}
	return r.ExitCode != nil && *r.ExitCode == 0 && r.Error == "" && rStdout == stdout
}
