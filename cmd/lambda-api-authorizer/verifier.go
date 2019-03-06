package main

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2/jwt"
)

const (
	kGoogleIssuer = "accounts.google.com"
)

type TokenVerifier struct {
	keyProvider KeyProvider
	expIssuer   string
	expAudience string
}

func NewTokenVerifier(keyProvider KeyProvider, iss string, aud string) *TokenVerifier {
	return &TokenVerifier{
		keyProvider: keyProvider,
		expIssuer:   iss,
		expAudience: aud,
	}
}

func NewGoogleTokenVerifier(aud string) *TokenVerifier {
	return NewTokenVerifier(NewGoogleKeyProvider(), kGoogleIssuer, aud)
}

func (v *TokenVerifier) VerifyToken(ctx context.Context, token string, extra interface{}) (*jwt.Claims, error) {
	j, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var claims *jwt.Claims
	if err := j.Claims(v.keyProvider.Keys(ctx), &claims, extra); err == nil {
		if err := claims.Validate(jwt.Expected{
			Issuer:   v.expIssuer,
			Audience: []string{v.expAudience},
			Time:     time.Now().UTC(),
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		return claims, nil
	}

	return nil, errors.New("no key found to verify JWT token")
}
