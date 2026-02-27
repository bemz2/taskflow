package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

type tokenPayload struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (s *TokenService) Issue(userID uuid.UUID) (string, error) {
	header, err := s.encode(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payload, err := s.encode(tokenPayload{
		Sub: userID.String(),
		Exp: time.Now().Add(s.ttl).Unix(),
	})
	if err != nil {
		return "", err
	}

	signingInput := header + "." + payload
	signature := s.sign(signingInput)

	return signingInput + "." + signature, nil
}

func (s *TokenService) Parse(token string) (uuid.UUID, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return uuid.Nil, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(s.sign(signingInput))) {
		return uuid.Nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	if time.Now().Unix() >= payload.Exp {
		return uuid.Nil, ErrTokenExpired
	}

	userID, err := uuid.Parse(payload.Sub)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

func (s *TokenService) encode(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal token part: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func (s *TokenService) sign(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
