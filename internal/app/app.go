// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/r-mol/Test_Service/internal/api"
	"github.com/r-mol/Test_Service/internal/config"
	"github.com/r-mol/Test_Service/internal/db/repository"
	"github.com/r-mol/Test_Service/internal/usecase/auth"
	"github.com/r-mol/Test_Service/pkg/mail"
	"github.com/r-mol/Test_Service/pkg/pg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context, configPath string, jwtKey string) error {
	config, err := config.ParseConfig(configPath)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	pg, err := pg.NewClient(ctx, config.PG)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pg.Close()

	var mailer *mail.Mailer
	if config.Mailer != nil {
		mailer = mail.NewMailer(config.Mailer)
	}

	// init repo
	repo := repository.NewTokenRepo(pg)

	// init usecases
	authUseCase := auth.NewUseCase(jwtKey, repo, mailer)

	// init routes
	authRoute := api.NewAuthAPIService(authUseCase)

	// init tg menu
	router := chi.NewRouter()
	router.Post("/tokens/{user_id}", authRoute.IssueTokensHandler)
	router.Post("/tokens/refresh", authRoute.RefreshTokenHandler)

	address := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)

	log.Infof("Starting server on %s", address)
	log.Fatal(http.ListenAndServe(address, router))

	return nil
}

func GetApp() *cobra.Command {
	var configPath string
	var jwtKey string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start test service",
		Run: func(cmd *cobra.Command, args []string) {
			err := Run(context.Background(), configPath, jwtKey)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "./config.yaml", "path to config file")
	cmd.Flags().StringVar(&jwtKey, "jwt_key", "", "jwt key")

	//_ = cmd.MarkFlagRequired("config")
	//_ = cmd.MarkFlagRequired("jwt_key")

	return cmd
}
