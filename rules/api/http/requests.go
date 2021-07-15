// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import "github.com/mainflux/mainflux/rules"

type streamReq struct {
	token  string
	stream rules.Stream
}

func (req streamReq) validate() error {
	if req.token == "" {
		return rules.ErrUnauthorizedAccess
	}
	if req.stream.Name == "" {
		return rules.ErrMalformedEntity
	}
	if req.stream.Row == "" {
		return rules.ErrMalformedEntity
	}
	if req.stream.Channel == "" {
		return rules.ErrMalformedEntity
	}
	return nil
}

type listReq struct {
	token string
}

func (req listReq) validate() error {
	if req.token == "" {
		return rules.ErrUnauthorizedAccess
	}
	return nil
}

type viewReq struct {
	token string
	name  string
}

func (req viewReq) validate() error {
	if req.token == "" {
		return rules.ErrUnauthorizedAccess
	}
	if req.name == "" {
		return rules.ErrMalformedEntity
	}
	return nil
}

type ruleReq struct {
	token          string
	ID             string `json:"id"`
	SQL            string `json:"sql"`
	Host           string `json:"host"`
	Port           string `json:"port"`
	Channel        string `json:"channel"`
	Subtopic       string `json:"subtopic"`
	SendMetaToSink bool   `json:"send_meta_to_sink"`
}

func (req ruleReq) validate() error {
	if req.token == "" {
		return rules.ErrUnauthorizedAccess
	}
	if req.ID == "" {
		return rules.ErrMalformedEntity
	}
	if req.SQL == "" {
		return rules.ErrMalformedEntity
	}
	if req.Channel == "" {
		return rules.ErrMalformedEntity
	}
	return nil
}

type controlReq struct {
	token  string
	id     string
	action string
}

func (req controlReq) validate() error {
	if req.token == "" {
		return rules.ErrUnauthorizedAccess
	}
	if req.id == "" {
		return rules.ErrMalformedEntity
	}
	if req.action == "" {
		return rules.ErrMalformedEntity
	}
	return nil
}

type deleteReq struct {
	token      string
	name       string
	kuiperType string
}

func (req deleteReq) validate() error {
	if req.token == "" {
		return rules.ErrUnauthorizedAccess
	}
	if req.name == "" {
		return rules.ErrMalformedEntity
	}
	if req.kuiperType == "" {
		return rules.ErrMalformedEntity
	}
	return nil
}
