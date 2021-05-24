// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package readers

import "errors"

const (
	// EqualKey represents the equal comparison operator key.
	EqualKey = "eq"
	// LowerThanKey represents the lower-than comparison operator key.
	LowerThanKey = "lt"
	// LowerThanEqualKey represents the lower-than-or-equal comparison operator key.
	LowerThanEqualKey = "le"
	// GreaterThanKey represents the greater-than-or-equal comparison operator key.
	GreaterThanKey = "gt"
	// GreaterThanEqualKey represents the greater-than-or-equal comparison operator key.
	GreaterThanEqualKey = "ge"
)

// ErrNotFound indicates that requested entity doesn't exist.
var ErrNotFound = errors.New("entity not found")

// MessageRepository specifies message reader API.
type MessageRepository interface {
	// ReadAll skips given number of messages for given channel and returns next
	// limited number of messages.
	ReadAll(pm PageMetadata) (MessagesPage, error)
}

// Message represents any message format.
type Message interface{}

// MessagesPage contains page related metadata as well as list of messages that
// belong to this page.
type MessagesPage struct {
	PageMetadata
	Total    uint64
	Messages []Message
}

// PageMetadata represents the parameters used to create database queries
type PageMetadata struct {
	ChanID    string `json:"channel_id"`
	Offset    uint64 `json:"offset"`
	Limit     uint64 `json:"limit"`
	Subtopic  string `json:"subtopic,omitempty"`
	Publisher string `json:"publisher,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Format    string `json:"format,omitempty"`
	// Name        string                 `json:"name,omitempty"`
	// Value       float64                `json:"v,omitempty"`
	// Comparator  string                 `json:"comparator,omitempty"`
	// BoolValue   bool                   `json:"vb,omitempty"`
	// StringValue string                 `json:"vs,omitempty"`
	// DataValue   string                 `json:"vd,omitempty"`
	// From        float64                `json:"from,omitempty"`
	// To          float64                `json:"to,omitempty"`
	Query map[string]interface{} `json:"query,omitempty"`
}

// type PageMetadata map[string]interface{}s

// ParseValueComparator convert comparison operator keys into mathematic anotation
func ParseValueComparator(query map[string]interface{}) string {
	comparator := "="
	val, ok := query["comparator"]
	if ok {
		switch val.(string) {
		case EqualKey:
			comparator = "="
		case LowerThanKey:
			comparator = "<"
		case LowerThanEqualKey:
			comparator = "<="
		case GreaterThanKey:
			comparator = ">"
		case GreaterThanEqualKey:
			comparator = ">="
		}
	}

	return comparator
}
