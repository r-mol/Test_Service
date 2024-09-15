package mail

import "golang.org/x/xerrors"

type Config struct {
	SmtpAddress string
	SmtpPort    int
	AuthorName  string
	AuthorPwd   string
}

func (cfg *Config) ValidateConfig() error {
	switch {
	case cfg.SmtpAddress == "":
		return xerrors.New("\"smtp_address\" is required")
	case cfg.SmtpPort == 0:
		return xerrors.New("\"smtp_port\" is required")
	case cfg.AuthorName == "":
		return xerrors.New("\"author_name\" is required")
	case cfg.AuthorPwd == "":
		return xerrors.New("\"author_pwd\" is required")
	}

	return nil
}
