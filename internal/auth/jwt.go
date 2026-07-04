package auth

import (
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

// JWTClaims are the custom claims carried by every token.
type JWTClaims struct {
	JWTKind JWTKind `json:"kind"`
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
// subject is the user's email; for a service token it is the service account id.
func (m *JWTManager) Generate(subject string, kind JWTKind) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		JWTKind: kind,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
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
