package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("o token de acesso expirou")
	ErrInvalidToken = errors.New("token de acesso inválido ou corrompido")
)

type Service struct {
	secretKey []byte
	tokenDuration time.Duration
}

func NewJWTService(secretKey string, tokenDuration time.Duration) *Service {
	return &Service{
		secretKey: []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

func (s *Service) GenerateToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
			Subject: userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)), // token expira em 1 hora
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("falha ao assinar o token: %w", err)
	}
	return tokenString, nil
}

func (s *Service) ValidateToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
	
}