package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug  bool   `envconfig:"debug" default:"false"`
	Host   string `envconfig:"host" default:"localhost"`
	Port   string `envconfig:"port" default:"8080"`
	TTLSec int    `envconfig:"ttl_seconds" default:"1800"` // default is 30 minutes
}

func NewWithName(serviceName string) (Config, error) {
	var s Config
	err := envconfig.Process(serviceName, &s)
	if err != nil {
		return s, err
	}

	return s, nil
}

func New() (Config, error) {
	return NewWithName("")
}
