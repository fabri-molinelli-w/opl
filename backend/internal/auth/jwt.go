package auth

import (
	"fmt"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

// Claims are the JWT claims described in RN-2.
type Claims struct {
	Email string      `json:"email"`
	Role  domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// TokenIssuer signs and verifies HS256 tokens.
type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl}
}

// Issue signs a token for the given user.
func (t *TokenIssuer) Issue(user domain.User) (string, error) {
	now := time.Now()
	claims := Claims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// Parse validates a token and returns its claims. Any failure is reported as
// ErrUnauthenticated so callers never leak the reason.
func (t *TokenIssuer) Parse(raw string) (Claims, error) {
	claims := Claims{}
	_, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrUnauthenticated
		}
		return t.secret, nil
	})
	if err != nil {
		return Claims{}, domain.ErrUnauthenticated
	}
	return claims, nil
}
