package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Hostport       string `yaml:"hostport"`
	DataSourceName string `yaml:"dataSourceName"`
}

func ReadConfig(filename string) (*Config, error) {
	config := &Config{}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config file %s err=%w", filename, err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("invalid config, err=%w", err)
	}

	return &Config{}, nil
}
