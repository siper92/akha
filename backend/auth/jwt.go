package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuerName = "akha"

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Keyring interface {
	TokenIssuer
	TokenVerifier
}

type claims struct {
	Tier Tier `json:"tier"`
	jwt.RegisteredClaims
}

type keyring struct {
	kp  KeyPair
	ttl time.Duration
}

var _ Keyring = (*keyring)(nil)

func NewKeyring(kp KeyPair, ttl time.Duration) Keyring {
	return &keyring{kp: kp, ttl: ttl}
}

func (k *keyring) Issue(ctx context.Context, subject string, tier Tier) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(k.ttl)
	id := make([]byte, 8)
	if _, err := rand.Read(id); err != nil {
		return "", time.Time{}, err
	}
	c := claims{
		Tier: tier,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerName,
			Subject:   subject,
			ID:        hex.EncodeToString(id),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c).SignedString(k.kp.Private)
	if err != nil {
		return "", time.Time{}, err
	}
	return tok, exp, nil
}

func (k *keyring) Verify(ctx context.Context, token string) (Claims, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, k.key,
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithIssuer(issuerName),
		jwt.WithExpirationRequired(),
	)
	if errors.Is(err, jwt.ErrTokenExpired) {
		return Claims{}, ErrExpiredToken
	}
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return Claims{
		Subject:   c.Subject,
		Tier:      c.Tier,
		IssuedAt:  c.IssuedAt.Time,
		ExpiresAt: c.ExpiresAt.Time,
	}, nil
}

func (k *keyring) key(t *jwt.Token) (any, error) {
	return k.kp.Public, nil
}
