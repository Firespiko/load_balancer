package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type BackendConfig struct {
	Url        string `yaml:"url"`
	Weight     int    `yaml:"weight"`
	HealthPath string `yaml:"health_path"`
}

type Config struct {
	Port           int64           `yaml:"port"`
	HealthInterval time.Duration   `yaml:"health_interval"`
	Algorithm      string          `yaml:"algorithm"`
	Backends       []BackendConfig `yaml:"backends"`
	RequestTimeout time.Duration   `yaml:"request_timeout"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", config)
	return &config, nil
}
