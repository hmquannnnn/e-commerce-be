package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewJWTManager(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:            []byte(secretKey),
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (m *JWTManager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID: userID.String(),
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "user-service",
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)

	return tokenString, err
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return m.secretKey, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*JWTClaims)

	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *JWTManager) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)
	return token, nil
}

func (m *JWTManager) GetAccessTokenDuration() time.Duration {
	return m.accessTokenDuration
}

func (m *JWTManager) GetRefreshTokenDuration() time.Duration {
	return m.refreshTokenDuration
}
