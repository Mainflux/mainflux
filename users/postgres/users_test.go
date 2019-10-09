// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSave(t *testing.T) {
	email := "user-save@example.com"

	cases := []struct {
		desc string
		user users.User
		err  error
	}{
		{
			desc: "new user",
			user: users.User{
				Email:    email,
				Password: "pass",
			},
			err: errors.New(""),
		},
		{
			desc: "duplicate user",
			user: users.User{
				Email:    email,
				Password: "pass",
			},
			err: users.ErrConflict,
		},
	}

	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.New(dbMiddleware)

	for _, tc := range cases {
		err := repo.Save(context.Background(), tc.user)
		switch v := err.(type) {
		case errors.Error:
			assert.True(t, v.Contains(tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		default:
			assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		}
	}
}

func TestSingleUserRetrieval(t *testing.T) {
	email := "user-retrieval@example.com"

	dbMiddleware := postgres.NewDatabase(db)
	repo := postgres.New(dbMiddleware)
	err := repo.Save(context.Background(), users.User{
		Email:    email,
		Password: "pass",
	})
	require.True(t, err.IsEmpty(), fmt.Sprintf("unexpected error: %s", err))

	cases := map[string]struct {
		email string
		err   error
	}{
		"existing user":     {email, errors.New("")},
		"non-existing user": {"unknown@example.com", users.ErrNotFound},
	}

	for desc, tc := range cases {
		_, err := repo.RetrieveByID(context.Background(), tc.email)
		switch v := err.(type) {
		case errors.Error:
			assert.True(t, v.Contains(tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		default:
			assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		}
	}
}
