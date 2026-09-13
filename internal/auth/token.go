package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

type TokenManager struct {
	secret             []byte
	tokenExpireSeconds int64
}

func NewTokenManager(secret string, tokenExpireSeconds int64) *TokenManager {
	return &TokenManager{
		secret:             []byte(secret),
		tokenExpireSeconds: tokenExpireSeconds,
	}
}

func (m *TokenManager) Issue(userID uint64) (string, error) {
	expiresAt := time.Now().Add(time.Duration(m.tokenExpireSeconds) * time.Second).Unix()
	payload := strconv.FormatUint(userID, 10) + ":" + strconv.FormatInt(expiresAt, 10)
	signature := m.sign(payload)

	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signature, nil
}

func (m *TokenManager) Parse(token string) (uint64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, ErrUnauthorized
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, ErrUnauthorized
	}

	payload := string(payloadBytes)
	if !hmac.Equal([]byte(m.sign(payload)), []byte(parts[1])) {
		return 0, ErrUnauthorized
	}

	fields := strings.Split(payload, ":")
	if len(fields) != 2 {
		return 0, ErrUnauthorized
	}

	userID, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, ErrUnauthorized
	}

	expiresAt, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, ErrUnauthorized
	}
	if time.Now().Unix() >= expiresAt {
		return 0, ErrUnauthorized
	}

	return userID, nil
}

func (m *TokenManager) ExpiresIn() int64 {
	return m.tokenExpireSeconds
}

func (m *TokenManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
