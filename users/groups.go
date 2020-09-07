// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
)

// User represents a Mainflux user account. Each user is identified given its
// email and password.
type Group struct {
	ID          string
	Name        string
	Owner       User
	Parent      *Group
	Description string
	Attributes  map[string]interface{}
	//Policies   map[string]Policy
	Metadata map[string]interface{}
}

// GroupRepository specifies an group persistence API.
type GroupRepository interface {
	// Save persists the group.
	Save(ctx context.Context, g Group) (Group, error)

	// Update updates the group data.
	Update(ctx context.Context, g Group) error

	// Delete deletes group for given id
	Delete(ctx context.Context, id string) error

	// RetrieveByID retrieves group by its unique identifier.
	RetrieveByID(ctx context.Context, id string) (Group, error)

	// RetrieveByName retrieves group by name
	RetrieveByName(ctx context.Context, name string) (Group, error)

	// RetrieveAll retrieves a group subtree created by owner starting from group with groupName
	RetrieveAll(ctx context.Context, ownerID, groupName string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// RetrieveAllForUser retrieves all groups that user belongs to
	RetrieveAllForUser(ctx context.Context, userID string, offset, limit uint64, gm Metadata) (GroupPage, error)

	// AssignUser adds user to group.
	AssignUser(ctx context.Context, userID, groupID string) error

	// RemoveUser
	RemoveUser(ctx context.Context, userID, groupID string) error
}
