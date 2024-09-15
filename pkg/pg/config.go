package pg

import (
	"fmt"

	"golang.org/x/xerrors"
)

type Config struct {
	Hosts    []string `yaml:"hosts"`
	Port     int      `yaml:"port"`
	User     string   `yaml:"user"`
	Password string   `yaml:"password"`
	DBName   string   `yaml:"db_name"`
	SSLMode  string   `yaml:"ssl_mode"`
	MaxConn  uint     `yaml:"max_conn"`
}

func (cfg *Config) ValidateConfig() error {
	switch {
	case cfg.Hosts == nil:
		return xerrors.New("\"hosts\" is required")
	case cfg.Port == 0:
		return xerrors.New("\"port\" is required")
	case cfg.User == "":
		return xerrors.New("\"user\" is required")
	case cfg.Password == "":
		return xerrors.New("\"password\" is required")
	case cfg.DBName == "":
		return xerrors.New("\"db_name\" is required")
	}

	return nil
}

func (cfg *Config) MakeConnStrings() map[string]string {
	res := map[string]string{}

	connString := "host=%s port=%d user=%s password=%s dbname=%s"
	if cfg.SSLMode != "" {
		connString += " sslmode=" + cfg.SSLMode
	}

	for _, host := range cfg.Hosts {
		res[host] = fmt.Sprintf(connString, host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	}

	return res
}
