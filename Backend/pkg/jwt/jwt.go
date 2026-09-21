package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	Subject   string
	ExpiresAt time.Time
	TokenType TokenType
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) (*Manager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 characters")
	}
	if accessTTL <= 0 || refreshTTL <= 0 {
		return nil, errors.New("JWT access and refresh TTLs must be positive")
	}
	return &Manager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}, nil
}

func (m *Manager) GeneratePair(subject string) (TokenPair, error) {
	accessToken, err := m.generate(subject, AccessToken, m.accessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, err := m.generate(subject, RefreshToken, m.refreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (m *Manager) generate(subject string, tokenType TokenType, ttl time.Duration) (string, error) {
	if subject == "" {
		return "", errors.New("JWT subject is required")
	}
	now := time.Now().UTC()
	claims := jwtv5.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
		"typ": string(tokenType),
	}
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) Parse(tokenString string, expectedType TokenType) (Claims, error) {
	claims := jwtv5.MapClaims{}
	token, err := jwtv5.ParseWithClaims(tokenString, claims, func(token *jwtv5.Token) (any, error) {
		if token.Method != jwtv5.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	subject, ok := claims["sub"].(string)
	tokenType, okType := claims["typ"].(string)
	expiresAt, expiryErr := claims.GetExpirationTime()
	if !ok || !okType || expiryErr != nil || subject == "" || TokenType(tokenType) != expectedType || expiresAt == nil {
		return Claims{}, ErrInvalidToken
	}
	return Claims{Subject: subject, ExpiresAt: expiresAt.Time, TokenType: TokenType(tokenType)}, nil
}
