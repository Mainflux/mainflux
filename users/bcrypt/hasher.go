//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0

// Package bcrypt provides a hasher implementation utilising bcrypt.
package bcrypt

import (
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/users"
	"golang.org/x/crypto/bcrypt"
)

const cost int = 10

var _ users.Hasher = (*bcryptHasher)(nil)

type bcryptHasher struct{}

// New instantiates a bcrypt-based hasher implementation.
func New() users.Hasher {
	return &bcryptHasher{}
}

func (bh *bcryptHasher) Hash(pwd string) (string, errors.Error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), cost)
	if err != nil {
		return "", errors.Cast(err)
	}

	return string(hash), errors.New("")
}

func (bh *bcryptHasher) Compare(plain, hashed string) errors.Error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return errors.Cast(err)
}
