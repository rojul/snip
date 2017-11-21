package api

import (
	"encoding/json"
	"testing"

	"github.com/rojul/snip/api/runner"
)

func TestLanguagesHelloWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var ls []*Language
	if *testLanguages != true {
		t.Log("pass -languages to test all languages")
		ls = []*Language{mustGetAsh()}
	} else {
		ls = testH.GetLanguages()
	}

	for _, l := range ls {
		t.Run(l.ID, func(t *testing.T) { testLanguageHelloWorld(t, l) })
	}
}

func testLanguageHelloWorld(t *testing.T, l *Language) {
	if l.NotRunnable {
		t.Skip("not runnable")
	}
	p := defaultPayload(l, l.HelloWorld)
	expected := resultFromMessage("Hello World\n")
	actual, err := testH.runContainerSync(p, l)
	if err != nil {
		t.Fatal(err)
	}
	expectedJSON := mustToJSON(expected)
	actualJSON := mustToJSON(actual)
	if expectedJSON != actualJSON {
		t.Fatalf("\nexpected: %s\nactual:   %s", expectedJSON, actualJSON)
	}
	t.Logf("ok: %s", actualJSON)
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

func defaultPayload(l *Language, content string) *Payload {
	return &Payload{
		Language: l.ID,
		Payload: runner.Payload{
			Files: []*runner.File{
				{
					Name:    "main." + l.Extension,
					Content: content,
				},
			},
		},
	}
}

func resultFromMessage(message string) *runner.Result {
	n := 0
	return &runner.Result{
		Events: []*runner.Event{
			{
				Message: message,
				Type:    "stdout",
			},
		},
		ExitCode: &n,
	}
}
