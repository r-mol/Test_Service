package config

import "golang.org/x/xerrors"

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func validateServerConfig(config *ServerConfig) error {
	switch {
	case config.Host == "":
		return xerrors.New("\"host\" is required")
	case config.Port == "":
		return xerrors.New("\"port\" is required")
	}

	return nil
}
