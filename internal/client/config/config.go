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
)

type Config struct {
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
	Hostport string `yaml:"hostport"`
}

func ReadConfig(filename string) (*Config, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return utils.ReadYaml[Config](filepath.Join(dir, configPath, filename))
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

	if login, ok := params["login"]; ok {
		config.Login = login
	}

	if password, ok := params["password"]; ok {
		config.Password = password
	}

	if hostport, ok := params["hostport"]; ok {
		config.Hostport = hostport
	}

	return utils.WriteYaml(fullPath, config)
}
