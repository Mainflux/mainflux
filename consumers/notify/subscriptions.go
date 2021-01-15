// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package notify

import "context"

// Subscription represents a user Subscription.
type Subscription struct {
	ID      string
	OwnerID string
	Contact string
	Topic   string
}

// SubscriptionPage represents page metadata with content.
type SubscriptionPage struct {
	PageMetadata
	Subscriptions []Subscription
}

// PageMetadata contains page metadata that helps navigation.
type PageMetadata struct {
	Total  uint
	Offset uint
	// Limit values less than 0 indicate no limit.
	Limit   int
	Topic   string
	Contact string
}

// SubscriptionsRepository specifies a Subscription persistence API.
type SubscriptionsRepository interface {
	// Save persists a subscription. Successful operation is indicated by non-nil
	// error response.
	Save(ctx context.Context, sub Subscription) (string, error)

	// Retrieve retrieves the subscription for the given id.
	Retrieve(ctx context.Context, id string) (Subscription, error)

	// RetrieveAll retrieves all the subscriptions for the given metadata.
	RetrieveAll(ctx context.Context, pm PageMetadata) ([]Subscription, error)

	// Remove removes the subscription having the provided an ID.
	Remove(ctx context.Context, id string) error
}
