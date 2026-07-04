package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTKind distinguishes a user session token from a service account token.
type JWTKind string

const (
	JWTKindUser    JWTKind = "user"
	JWTKindService JWTKind = "service"
)

// JWTClaims are the custom claims carried by every token. Labels are the
// key/value pairs a user attaches to a service account to identify its token.
type JWTClaims struct {
	JWTKind JWTKind           `json:"kind"`
	Labels  map[string]string `json:"labels,omitempty"`
	Salt    string            `json:"slt,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager signs and parses tokens with a shared HMAC secret.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager builds a Manager. ttl is how long issued tokens remain valid.
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), ttl: ttl}
}

// TTL returns how long issued tokens remain valid; callers use it to expire
// cached copies alongside the token.
func (m *JWTManager) TTL() time.Duration {
	return m.ttl
}

// Generate signs a token for the given subject and kind. For a user token the
// subject is the user's email; for a service token it is the service account id
// and labels carries the account's key/value pairs. labels may be nil.
func (m *JWTManager) Generate(subject string, kind JWTKind, labels map[string]string) (string, error) {
	return m.GenerateWithTTL(subject, kind, labels, m.ttl)
}

// GenerateWithTTL is like Generate but uses an explicit lifetime, letting a
// caller override the default TTL per token (e.g. a per-account TTL).
func (m *JWTManager) GenerateWithTTL(subject string, kind JWTKind, labels map[string]string, ttl time.Duration) (string, error) {
	salt, err := randomSalt()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := JWTClaims{
		JWTKind: kind,
		Labels:  labels,
		Salt:    salt,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// randomSalt returns a cryptographically-random hex string used to make every
// generated token unique.
func randomSalt() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// Parse validates a signed token and returns its claims.
func (m *JWTManager) Parse(raw string) (*JWTClaims, error) {
	claims := &JWTClaims{}

	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}
