package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Manager struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	Subject   string
	ExpiresAt time.Time
}

func NewManager(secret string, ttl time.Duration) (*Manager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 characters")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT TTL must be positive")
	}
	return &Manager{secret: []byte(secret), ttl: ttl}, nil
}

func (m *Manager) Generate(subject string) (string, error) {
	if subject == "" {
		return "", errors.New("JWT subject is required")
	}
	now := time.Now().UTC()
	claims := jwtv5.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(now.Add(m.ttl)),
	}
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) Parse(tokenString string) (Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &jwtv5.RegisteredClaims{}, func(token *jwtv5.Token) (any, error) {
		if token.Method != jwtv5.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	parsed, ok := token.Claims.(*jwtv5.RegisteredClaims)
	if !ok || parsed.Subject == "" || parsed.ExpiresAt == nil {
		return Claims{}, ErrInvalidToken
	}
	return Claims{Subject: parsed.Subject, ExpiresAt: parsed.ExpiresAt.Time}, nil
}
