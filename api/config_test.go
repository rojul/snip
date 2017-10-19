package api

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/docker/go-units"
)

func TestEnvConfig(t *testing.T) {
	var envTests = []struct {
		envKey   string
		envVal   string
		field    string
		expected interface{}
	}{
		{"RUN_TIMEOUT", "5s", "RunTimeout", 5 * time.Second},
		{"MEMORY", "5m", "Memory", 5 * int64(units.MiB)},
		{"JSON_LOGGING", "true", "JSONLogging", true},
		{"SNIPPET_SIZE_LIMIT", "5k", "SnippetSizeLimit", 5 * int64(units.KB)},
	}

	for _, tt := range envTests {
		os.Setenv("SNIP_"+tt.envKey, tt.envVal)

		c, err := configFromEnv()
		if err != nil {
			t.Fatal(err)
		}

		actual := reflect.ValueOf(c).Elem().FieldByName(tt.field).Interface()
		if actual != tt.expected {
			t.Errorf("field %s: expected %#v, actual %#v", tt.field, tt.expected, actual)
		}
	}
}
