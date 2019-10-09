// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package jwt provides a JWT identity provider.
package jwt

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/users"
)

const (
	issuer   string        = "mainflux"
	duration time.Duration = 10 * time.Hour
)

var _ users.IdentityProvider = (*jwtIdentityProvider)(nil)

type jwtIdentityProvider struct {
	secret string
}

// New instantiates a JWT identity provider.
func New(secret string) users.IdentityProvider {
	return &jwtIdentityProvider{secret}
}

func (idp *jwtIdentityProvider) TemporaryKey(id string) (string, errors.Error) {
	now := time.Now().UTC()
	exp := now.Add(duration)

	claims := jwt.StandardClaims{
		Subject:   id,
		Issuer:    issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: exp.Unix(),
	}

	return idp.jwt(claims)
}

func (idp *jwtIdentityProvider) Identity(key string) (string, errors.Error) {
	token, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, users.ErrUnauthorizedAccess
		}

		return []byte(idp.secret), errors.Empty()
	})

	if err != nil {
		return "", errors.Wrap(users.ErrUnauthorizedAccess, errors.Cast(err))

	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub := claims["sub"]; sub != nil {
			return sub.(string), nil
		}
	}

	return "", users.ErrUnauthorizedAccess
}

func (idp *jwtIdentityProvider) jwt(claims jwt.StandardClaims) (string, errors.Error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := token.SignedString([]byte(idp.secret))
	return tok, errors.Cast(err)
}
