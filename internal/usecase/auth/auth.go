package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/r-mol/Test_Service/internal/domain"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	durationAccessToken  = 15 * time.Minute
	durationRefreshToken = 24 * time.Hour
)

type TokenRepo interface {
	GetByUserID(ctx context.Context, user_id uuid.UUID) (*domain.Token, error)
	UpdateOrCreate(ctx context.Context, token *domain.Token) (int, error)
}

type MailClient interface {
	SendMail(From string, To string, Subject string, Body string) error
}

type UseCase struct {
	jwtKey     string
	tokenRepo  TokenRepo
	mailClient MailClient
}

func NewUseCase(
	jwtKey string,
	tokenRepo TokenRepo,
	mailClient MailClient,
) *UseCase {
	return &UseCase{
		jwtKey:     jwtKey,
		tokenRepo:  tokenRepo,
		mailClient: mailClient,
	}
}

func (uc *UseCase) GetNewTokens(
	ctx context.Context,
	userID string,
	ip string,
) (string, string, error) {
	accessToken, err := uc.generateJWT(userID, ip)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := uc.generateRefreshToken(userID, ip)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	hashToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("generate hash token: %w", err)
	}

	_, err = uc.tokenRepo.UpdateOrCreate(ctx, &domain.Token{
		UserID: uuid.MustParse(userID),
		Hash:   string(hashToken),
	})
	if err != nil {
		return "", "", fmt.Errorf("update of create token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (uc *UseCase) RefreshTokens(
	ctx context.Context,
	refreshToken string,
	ip string,
) (string, string, error) {
	claims, err := uc.parseToken(refreshToken)
	if err != nil {
		return "", "", &domain.UnauthorizedError{Message: "Invalid refresh token"}
	}

	token, err := uc.tokenRepo.GetByUserID(ctx, uuid.MustParse(claims.UserID))
	if err != nil {
		return "", "", fmt.Errorf("get refresh token by user_id: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(token.Hash), []byte(refreshToken))
	if err != nil {
		return "", "", &domain.UnauthorizedError{Message: "Invalid refresh token"}
	}

	if claims.IP != ip {
		if uc.mailClient != nil {
			// Todo (rmolochko@): get user email from db
			err = uc.mailClient.SendMail(MailAddress, "", SubjectAnotherIPWarning, BodyAnotherIPWarning)
			if err != nil {
				log.Error("Unable to send mail about ip mismatch")
			}
		} else {
			log.Warn("Unauthorized - ip mismatch")
		}
	}

	newAccessToken, err := uc.generateJWT(claims.UserID, ip)
	if err != nil {
		return "", "", fmt.Errorf("generate new access token: %w", err)
	}

	newRefreshToken, err := uc.generateRefreshToken(claims.UserID, ip)
	if err != nil {
		return "", "", fmt.Errorf("generate new refresh token: %w", err)
	}

	hashToken, err := bcrypt.GenerateFromPassword([]byte(newRefreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("generate hash token: %w", err)
	}

	_, err = uc.tokenRepo.UpdateOrCreate(ctx, &domain.Token{
		UserID: uuid.MustParse(claims.UserID),
		Hash:   string(hashToken),
	})
	if err != nil {
		return "", "", fmt.Errorf("update refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (uc *UseCase) generateJWT(userID, ip string) (string, error) {
	expirationTime := time.Now().Add(durationAccessToken)
	claims := &domain.Claims{
		UserID: userID,
		IP:     ip,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(uc.jwtKey))
}

func (uc *UseCase) generateRefreshToken(userID, ip string) (string, error) {
	str := fmt.Sprintf("%s %s", userID, ip)
	return base64.StdEncoding.EncodeToString([]byte(str)), nil
}

func (uc *UseCase) parseToken(encodedToken string) (*domain.Claims, error) {
	data, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	strs := strings.Split(string(data), " ")
	if len(strs) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	return &domain.Claims{
		UserID: strs[0],
		IP:     strs[1],
	}, nil
}
