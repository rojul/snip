package api

import (
	"reflect"
	"time"

	units "github.com/docker/go-units"
	"github.com/spf13/viper"
)

type Config struct {
	RunTimeout         time.Duration `mapstructure:"RUN_TIMEOUT"`
	Memory             int64         `mapstructure:"MEMORY"`
	NanoCPUs           int64         `mapstructure:"NANO_CPUS"`
	CPUShares          int64         `mapstructure:"CPU_SHARES"`
	PidsLimit          int64         `mapstructure:"PIDS_LIMIT"`
	NetworkEnabled     bool          `mapstructure:"NETWORK_ENABLED"`
	MongoURL           string        `mapstructure:"MONGO_URL"`
	MongoDB            string        `mapstructure:"MONGO_DB"`
	JSONLogging        bool          `mapstructure:"JSON_LOGGING"`
	SnippetSizeLimit   int64         `mapstructure:"SNIPPET_SIZE_LIMIT"`
	ReturnSizeLimit    int64         `mapstructure:"RETURN_SIZE_LIMIT"`
	CorsEnabled        bool          `mapstructure:"CORS_ENABLED"`
	HTTPAddr           string        `mapstructure:"HTTP_ADDR"`
	DefaultImagePrefix string        `mapstructure:"DEFAULT_IMAGE_PREFIX"`
	LanguagesFile      string        `mapstructure:"LANGUAGES_FILE"`
}

func defaultConfig() *Config {
	return &Config{
		RunTimeout:         15 * time.Second,
		Memory:             512 * units.MiB,
		CPUShares:          64,
		PidsLimit:          35,
		SnippetSizeLimit:   1 * units.MiB,
		MongoURL:           "mongo",
		MongoDB:            "snip",
		DefaultImagePrefix: "snip",
		LanguagesFile:      "languages.yml",
		ReturnSizeLimit:    100 * units.KiB,
	}
}

func parseInt64WithUnit(v *viper.Viper, f func(string) (int64, error), key string) {
	s := v.GetString(key)
	if s == "" {
		return
	}
	n, err := f(s)
	if err != nil {
		return
	}
	v.Set(key, n)
}

func configFromEnv() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("snip")

	t := reflect.TypeOf(Config{})
	for i := 0; i < t.NumField(); i++ {
		v.BindEnv(t.Field(i).Tag.Get("mapstructure"))
	}

	parseInt64WithUnit(v, units.RAMInBytes, "memory")
	parseInt64WithUnit(v, units.FromHumanSize, "snippet_size_limit")
	parseInt64WithUnit(v, units.FromHumanSize, "return_size_limit")

	c := defaultConfig()
	err := v.Unmarshal(&c)
	return c, err
}
