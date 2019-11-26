// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/errors"
)

var (
	// ErrConflict indicates usage of the existing email during account
	// registration.
	ErrConflict = errors.New("email already taken")

	// ErrMalformedEntity indicates malformed entity specification
	// (e.g. invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request
	ErrNotFound = errors.New("non-existent entity")

	// ErrUserNotFound indicates a non-existent user request
	ErrUserNotFound = errors.New("non-existent user")

	// ErrScanMetadata indicates problem with metadata in db
	ErrScanMetadata = errors.New("Failed to scan metadata")

	// ErrMissingEmail indicates missing email for password reset request
	ErrMissingEmail = errors.New("missing email for password reset")

	// ErrMissingResetToken indicates malformed or missing reset token
	// for reseting password
	ErrMissingResetToken = errors.New("error missing reset token")

	// ErrGeneratingResetToken indicates error in generating password recovery
	// token
	ErrGeneratingResetToken = errors.New("error missing reset token")

	// ErrGetToken indicates error in getting signed token
	ErrGetToken = errors.New("Get signed token failed")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Register creates new user account. In case of the failed registration, a
	// non-nil error value is returned.
	Register(context.Context, User) errors.Error

	// Login authenticates the user given its credentials. Successful
	// authentication generates new access token. Failed invocations are
	// identified by the non-nil error values in the response.
	Login(context.Context, User) (string, errors.Error)

	// Get authenticated user info for the given token
	UserInfo(ctx context.Context, token string) (User, errors.Error)

	// UpdateUser updates the user metadata
	UpdateUser(ctx context.Context, token string, user User) errors.Error

	// GenerateResetToken email where mail will be sent.
	// host is used for generating reset link.
	GenerateResetToken(_ context.Context, email, host string) errors.Error

	// ChangePassword change users password for authenticated user.
	ChangePassword(_ context.Context, authToken, password, oldPassword string) errors.Error

	// ResetPassword change users password in reset flow.
	// token can be authentication token or password reset token.
	ResetPassword(_ context.Context, resetToken, password string) errors.Error

	//SendPasswordReset sends reset password link to email
	SendPasswordReset(_ context.Context, host, email, token string) errors.Error
}

var _ Service = (*usersService)(nil)

type usersService struct {
	users  UserRepository
	hasher Hasher
	email  Emailer
	auth   mainflux.AuthServiceClient
}

// New instantiates the users service implementation
func New(users UserRepository, hasher Hasher, auth mainflux.AuthServiceClient, m Emailer) Service {
	return &usersService{
		users:  users,
		hasher: hasher,
		auth:   auth,
		email:  m,
	}
}

func (svc usersService) Register(ctx context.Context, user User) errors.Error {
	hash, err := svc.hasher.Hash(user.Password)
	if err != nil {
		return errors.Wrap(ErrMalformedEntity, err)
	}

	user.Password = hash
	return svc.users.Save(ctx, user)
}

func (svc usersService) Login(ctx context.Context, user User) (string, errors.Error) {
	dbUser, err := svc.users.RetrieveByID(ctx, user.Email)
	if err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, err)
	}

	if err := svc.hasher.Compare(user.Password, dbUser.Password); err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return svc.issue(ctx, dbUser.Email, auth.LoginKey)
}

func (svc usersService) UserInfo(ctx context.Context, token string) (User, errors.Error) {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return User{}, err
	}

	dbUser, err := svc.users.RetrieveByID(ctx, id)
	if err != nil {
		return User{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return User{
		Email:    id,
		Password: "",
		Metadata: dbUser.Metadata,
	}, nil
}

func (svc usersService) UpdateUser(ctx context.Context, token string, u User) errors.Error {
	email, err := svc.identify(ctx, token)
	if err != nil || email != u.Email {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	user := User{
		Email:    email,
		Metadata: u.Metadata,
	}

	return svc.users.UpdateUser(ctx, user)
}

func (svc usersService) GenerateResetToken(ctx context.Context, email, host string) errors.Error {
	user, err := svc.users.RetrieveByID(ctx, email)
	if err != nil || user.Email == "" {
		return ErrUserNotFound
	}

	t, err := svc.issue(ctx, email, auth.ResetKey)
	if err != nil {
		return errors.Wrap(ErrGeneratingResetToken, err)
	}
	return svc.SendPasswordReset(ctx, host, email, t)
}

func (svc usersService) ResetPassword(ctx context.Context, resetToken, password string) errors.Error {
	email, err := svc.identify(ctx, resetToken)
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	u, err := svc.users.RetrieveByID(ctx, email)
	if err != nil || u.Email == "" {
		return ErrUserNotFound
	}

	password, err = svc.hasher.Hash(password)
	if err != nil {
		return err
	}
	return svc.users.UpdatePassword(ctx, email, password)
}

func (svc usersService) ChangePassword(ctx context.Context, authToken, password, oldPassword string) errors.Error {
	email, err := svc.identify(ctx, authToken)
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	u := User{
		Email:    email,
		Password: oldPassword,
	}
	if _, err := svc.Login(ctx, u); err != nil {
		return ErrUnauthorizedAccess
	}

	u, err = svc.users.RetrieveByID(ctx, email)
	if err != nil || u.Email == "" {
		return ErrUserNotFound
	}

	password, err = svc.hasher.Hash(password)
	if err != nil {
		return err
	}
	return svc.users.UpdatePassword(ctx, email, password)
}

// SendPasswordReset sends password recovery link to user
func (svc usersService) SendPasswordReset(_ context.Context, host, email, token string) errors.Error {
	to := []string{email}
	return svc.email.SendPasswordReset(to, host, token)
}

func (svc usersService) identify(ctx context.Context, token string) (string, errors.Error) {
	email, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, err)
	}
	return email.GetValue(), nil
}

func (svc usersService) issue(ctx context.Context, email string, keyType uint32) (string, errors.Error) {
	key, err := svc.auth.Issue(ctx, &mainflux.IssueReq{Issuer: email, Type: keyType})
	if err != nil {
		return "", errors.Wrap(ErrUserNotFound, err)
	}
	return key.GetValue(), nil
}
