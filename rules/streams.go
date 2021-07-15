// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rules

const (
	format     = "json"
	pluginType = "mainflux"
)

// Stream represents data used to create kuiper stream
type Stream struct {
	Name     string `json:"name,omitempty"`
	Row      string `json:"row"`
	Channel  string `json:"channel"`
	Subtopic string `json:"subtopic"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

// StreamInfo is used to fetch stream info from kuiper
type StreamInfo struct {
	Name         string `json:"Name"`
	StreamFields []struct {
		Name      string `json:"Name"`
		FieldType string `json:"FieldType"`
	} `json:"StreamFields"`
	Options struct {
		DATASOURCE string `json:"DATASOURCE"`
		FORMAT     string `json:"FORMAT"`
	} `json:"Options"`
}
