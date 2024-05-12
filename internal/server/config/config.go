package config

import (
	"github.com/kuzhukin/goph-keeper/internal/utils"
)

type Config struct {
	Hostport       string `yaml:"hostport"`
	DataSourceName string `yaml:"dataSourceName"`
}

func ReadConfig(filename string) (*Config, error) {
	return utils.ReadYaml[Config](filename)
}
