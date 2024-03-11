package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"os"
)

type CacheConfig struct {
	TTLSec                   int `envconfig:"ttl_seconds" default:"1800"`          // default is 30 minutes
	EvictionIntervalMilliSec int `envconfig:"eviction_interval_ms" default:"1000"` // default is 1 second
}
type Config struct {
	Debug bool        `envconfig:"debug" default:"false"`
	Host  string      `envconfig:"host" default:"0.0.0.0"`
	Port  string      `envconfig:"port" default:"8080"`
	Cache CacheConfig // default is 30 minutes
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
	var s Config
	serviceName := os.Getenv("SERVICE_NAME")
	err := envconfig.Process(serviceName, &s)
	if err != nil {
		return s, err
	}
	fmt.Printf("config loaded: %+v\n", s)
	return s, nil
}
