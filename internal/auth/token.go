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

const (
	TokenSubjectUser  = "user"
	TokenSubjectAdmin = "admin"
)

func NewTokenManager(secret string, tokenExpireSeconds int64) *TokenManager {
	return &TokenManager{
		secret:             []byte(secret),
		tokenExpireSeconds: tokenExpireSeconds,
	}
}

func (m *TokenManager) Issue(userID uint64) (string, error) {
	return m.IssueUser(userID)
}

func (m *TokenManager) IssueUser(userID uint64) (string, error) {
	return m.issue(TokenSubjectUser, userID)
}

func (m *TokenManager) IssueAdmin(adminID uint64) (string, error) {
	return m.issue(TokenSubjectAdmin, adminID)
}

func (m *TokenManager) issue(subject string, id uint64) (string, error) {
	expiresAt := time.Now().Add(time.Duration(m.tokenExpireSeconds) * time.Second).Unix()
	payload := subject + ":" + strconv.FormatUint(id, 10) + ":" + strconv.FormatInt(expiresAt, 10)
	signature := m.sign(payload)

	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signature, nil
}

func (m *TokenManager) Parse(token string) (uint64, error) {
	return m.ParseUser(token)
}

func (m *TokenManager) ParseUser(token string) (uint64, error) {
	return m.parse(token, TokenSubjectUser)
}

func (m *TokenManager) ParseAdmin(token string) (uint64, error) {
	return m.parse(token, TokenSubjectAdmin)
}

func (m *TokenManager) parse(token string, expectedSubject string) (uint64, error) {
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
	if len(fields) != 3 {
		return 0, ErrUnauthorized
	}
	if fields[0] != expectedSubject {
		return 0, ErrUnauthorized
	}

	id, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, ErrUnauthorized
	}

	expiresAt, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return 0, ErrUnauthorized
	}
	if time.Now().Unix() >= expiresAt {
		return 0, ErrUnauthorized
	}

	return id, nil
}

func (m *TokenManager) ExpiresIn() int64 {
	return m.tokenExpireSeconds
}

func (m *TokenManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
