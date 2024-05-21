package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/kuzhukin/goph-keeper/internal/utils"
)

const (
	configPath      = ".goph-keeper"
	hostportDefault = "http://localhost:34555"
	dbDefault       = "default.db"
)

type Config struct {
	Hostport string `yaml:"hostport"`
	Database string `yaml:"database"`
}

func ReadConfig(filename string) (*Config, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	confFilePath := filepath.Join(dir, configPath, filename)

	conf, err := utils.ReadYaml[Config](confFilePath)
	if err != nil {
		return nil, err
	}

	if conf.Hostport == "" {
		conf.Hostport = hostportDefault
	}

	if conf.Database == "" {
		conf.Database = dbDefault
	}

	return conf, nil
}

func UpdateConfig(filename string, params map[string]string) error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(dir, configPath, filename)

	config, err := ReadConfig(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			config = &Config{Hostport: hostportDefault}
		} else {
			return err
		}
	}

	if hostport, ok := params["hostport"]; ok {
		config.Hostport = hostport
	}

	if database, ok := params["database"]; ok {
		config.Database = database
	}

	return utils.WriteYaml(fullPath, config)
}
