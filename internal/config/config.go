package config

import (
	"fmt"
	"os"

	"github.com/r-mol/Test_Service/pkg/mail"
	"github.com/r-mol/Test_Service/pkg/pg"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server *ServerConfig `yaml:"server"`
	PG     *pg.Config    `yaml:"pg"`
	Mailer *mail.Config  `yaml:"mailer"`
}

func validateConfig(config *Config) error {
	switch {
	case config.Server == nil:
		return xerrors.New("\"server\" is required")
	case config.PG == nil:
		return xerrors.New("\"pg\" is required")
	}

	if err := validateServerConfig(config.Server); err != nil {
		return fmt.Errorf("validate server config: %w", err)
	}

	if err := config.PG.ValidateConfig(); err != nil {
		return fmt.Errorf("validate pg config: %w", err)
	}

	if config.Mailer != nil {
		if err := config.Mailer.ValidateConfig(); err != nil {
			return fmt.Errorf("validate mailer config: %w", err)
		}
	}

	return nil
}

func ParseConfig(path string) (*Config, error) {
	config := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("read file: %w", err)
	}

	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("umarshal config: %w", err)
	}

	if err = validateConfig(config); err != nil {
		return config, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}
