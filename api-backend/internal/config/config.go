package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DBConn        string `env:"DB_CONN"`
	CacheDisabled bool   `env:"CACHE_DISABLED" envDefault:"false"`
}

func Load() (Config, error) {
	return load(env.Options{})
}

func load(options env.Options) (Config, error) {
	options.FuncMap = map[reflect.Type]env.ParserFunc{
		reflect.TypeFor[bool](): parseBool,
	}

	cfg, err := env.ParseAsWithOptions[Config](options)
	if err != nil {
		return Config{}, err
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.CacheDisabled {
		return nil
	}

	if strings.TrimSpace(c.DBConn) == "" {
		return fmt.Errorf("DB_CONN is required when caching is enabled")
	}

	return nil
}

func parseBool(value string) (any, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yes", "on":
		return true, nil
	case "no", "off":
		return false, nil
	default:
		return strconv.ParseBool(value)
	}
}
