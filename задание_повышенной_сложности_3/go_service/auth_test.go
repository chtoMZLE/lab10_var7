package main

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken_Valid(t *testing.T) {
	token, err := GenerateToken("alice")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken_Valid(t *testing.T) {
	token, _ := GenerateToken("alice")
	claims, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, "alice", claims.Subject)
}

func TestParseToken_WrongSecret(t *testing.T) {
	// Подписываем другим секретом
	claims := Claims{
		Username: "hacker",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte("wrong-secret"))

	_, err := ParseToken(signed)
	assert.Error(t, err)
}

func TestParseToken_Expired(t *testing.T) {
	claims := Claims{
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(JWTSecret))

	_, err := ParseToken(signed)
	assert.Error(t, err)
}

func TestParseToken_Malformed(t *testing.T) {
	_, err := ParseToken("not.a.token")
	assert.Error(t, err)
}

func TestParseToken_WrongAlgorithm(t *testing.T) {
	// JWT с alg=none не должен проходить валидацию
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user","exp":9999999999}`))
	fakeToken := header + "." + payload + "."
	_, err := ParseToken(fakeToken)
	assert.Error(t, err)
}
