package service

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTokenServiceIssueAndParse(t *testing.T) {
	t.Parallel()

	svc := NewTokenService("test-secret", time.Minute)
	userID := uuid.New()

	token, err := svc.Issue(userID)

	require.NoError(t, err)
	require.Len(t, strings.Split(token, "."), 3)

	parsedUserID, err := svc.Parse(token)

	require.NoError(t, err)
	require.Equal(t, userID, parsedUserID)
}

func TestTokenServiceParseRejectsMalformedToken(t *testing.T) {
	t.Parallel()

	svc := NewTokenService("test-secret", time.Minute)

	_, err := svc.Parse("broken-token")

	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestTokenServiceParseRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	svc := NewTokenService("test-secret", -time.Second)
	token, err := svc.Issue(uuid.New())
	require.NoError(t, err)

	_, err = svc.Parse(token)

	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestTokenServiceParseRejectsTamperedToken(t *testing.T) {
	t.Parallel()

	svc := NewTokenService("test-secret", time.Minute)
	userID := uuid.New()
	token, err := svc.Issue(userID)
	require.NoError(t, err)

	tampered := token[:len(token)-1] + "x"

	_, err = svc.Parse(tampered)

	require.ErrorIs(t, err, ErrInvalidToken)
}
