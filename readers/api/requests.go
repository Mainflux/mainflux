// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers"
)

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	pageMeta readers.PageMetadata
}

type query struct {
	ChannelID string                 `json:"-" schema:"channel_id"`
	Offset    uint64                 `json:"-" schema:"offset"`
	Limit     uint64                 `json:"-" schema:"limit"`
	Subtopic  string                 `json:"subtopic,omitempty" schema:"subtopic"`
	Publisher string                 `json:"publisher,omitempty" schema:"publsher"`
	Protocol  string                 `json:"protocol,omitempty" schema:"protocol"`
	Format    string                 `json:"format,omitempty" schema:"format"`
	Query     map[string]interface{} `json:"query,omitempty"`
	// SenML-specific fields
	Comparator  string  `json:"comparator,omitempty" schema:"comparator"`
	Name        string  `json:"name,omitempty" schema:"name"`
	Unit        string  `json:"unit,omitempty" schema:"unit"`
	Value       float64 `json:"value,omitempty" schema:"value"`
	StringValue string  `json:"string_value,omitempty" schema:"string_value"`
	DataValue   string  `json:"data_value,omitempty" schema:"data_value"`
	BoolValue   bool    `json:"bool_value,omitempty" schema:"bool_value"`
	From        float64 `json:"from,omitempty" schema:"from"`
	To          float64 `json:"to,omitempty" schema:"to"`
}

func (q query) query() (map[string]interface{}, error) {
	data, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]interface{})
	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (q query) toPageMeta() (readers.PageMetadata, error) {
	m, err := q.query()
	if err != nil {
		return readers.PageMetadata{}, err
	}
	// ret := readers.PageMetadata{
	// 	// ChanID: q.ChannelID,
	// 	Offset: q.Offset,
	// 	Limit:  q.Limit,
	// 	// Subtopic:  q.Subtopic,
	// 	// Publisher: q.Publisher,
	// 	// Protocol:  q.Protocol,
	// 	// Format:    q.Format,
	// 	Query: m,
	// }
	return m, nil
}

func (req listMessagesReq) validate() error {
	limit, ok := req.pageMeta["limit"]
	if !ok {
		return errors.ErrInvalidQueryParams
	}
	if l, ok := limit.(int); !ok || l < 0 {
		return errors.ErrInvalidQueryParams
	}
	if req.pageMeta.Limit < 1 || req.pageMeta.Offset < 0 {
		return errors.ErrInvalidQueryParams
	}
	// if req.pageMeta.Comparator != "" &&
	// 	req.pageMeta.Comparator != readers.EqualKey &&
	// 	req.pageMeta.Comparator != readers.LowerThanKey &&
	// 	req.pageMeta.Comparator != readers.LowerThanEqualKey &&
	// 	req.pageMeta.Comparator != readers.GreaterThanKey &&
	// 	req.pageMeta.Comparator != readers.GreaterThanEqualKey {
	// 	return errors.ErrInvalidQueryParams
	// }

	return nil
}
