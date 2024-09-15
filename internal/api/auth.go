package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

type AuthUseCase interface {
	GetNewTokens(ctx context.Context, userID string, ip string) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string, ip string) (string, string, error)
}
type AuthAPIService struct {
	authUseCase AuthUseCase
}

func NewAuthAPIService(authUseCase AuthUseCase) *AuthAPIService {
	return &AuthAPIService{
		authUseCase: authUseCase,
	}
}

func (svc *AuthAPIService) IssueTokensHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("IssueTokensHandler", r)

	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr

	accessToken, refreshToken, err := svc.authUseCase.GetNewTokens(r.Context(), userID, ip)
	if err != nil {

		http.Error(w, fmt.Sprintf("unable to generate tokens: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (svc *AuthAPIService) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("RefreshTokenHandler", r)

	refreshToken := r.URL.Query().Get("refresh_token")
	if refreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr

	newAccessToken, newRefreshToken, err := svc.authUseCase.RefreshTokens(r.Context(), refreshToken, ip)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to refresh tokens", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}
