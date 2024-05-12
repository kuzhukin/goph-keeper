package client

import (
	"errors"
	"os"

	"github.com/kuzhukin/goph-keeper/internal/utils"
)

type AuthInfo struct {
	Login       string `yaml:"login"`
	Password    string `yaml:"password"`
	isRegistred bool
}

const clientConfigPath = "~/.gophkeeper/user.yaml"

func GetAuthInfo() (*AuthInfo, error) {
	info, err := utils.ReadYaml[AuthInfo](clientConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &AuthInfo{}, nil
		}

		return nil, err
	}

	return info, nil
}
