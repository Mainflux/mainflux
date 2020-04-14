// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// BoostrapConfig represents Configuration entity. It wraps information about external entity
// as well as info about corresponding Mainflux entities.
// MFThing represents corresponding Mainflux Thing ID.
// MFKey is key of corresponding Mainflux Thing.
// MFChannels is a list of Mainflux Channels corresponding Mainflux Thing connects to.
type BoostrapConfig struct {
	ThingID     string    `json:"thing_id,omitempty"`
	Channels    []string  `json:"channels,omitempty"`
	ExternalID  string    `json:"external_id,omitempty"`
	ExternalKey string    `json:"external_key,omitempty"`
	MFThing     string    `json:"mainflux_id,omitempty"`
	MFChannels  []Channel `json:"mainflux_channels,omitempty"`
	MFKey       string    `json:"mainflux_key,omitempty"`
	Name        string    `json:"name,omitempty"`
	ClientCert  string    `json:"client_cert,omitempty"`
	ClientKey   string    `json:"client_key,omitempty"`
	CACert      string    `json:"ca_cert,omitempty"`
	Content     string    `json:"content,omitempty"`
	State       int       `json:"state,omitempty"`
}

func (sdk mfSDK) AddBootstrap(key string, cfg BoostrapConfig) (string, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", ErrInvalidArgs
	}

	url := createURL(sdk.bootstrapURL, sdk.bootstrapPrefix, "configs")

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	resp, err := sdk.sendRequest(req, key, string(CTJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		if err := encodeError(resp.StatusCode); err != nil {
			return "", err
		}
		return "", ErrFailedCreation
	}

	id := strings.TrimPrefix(resp.Header.Get("Location"), "/things/configs/")
	return id, nil
}

func (sdk mfSDK) ViewBoostrap(key, id string) (BoostrapConfig, error) {
	endpoint := fmt.Sprintf("%s/%s", "configs", id)
	url := createURL(sdk.bootstrapURL, sdk.bootstrapPrefix, endpoint)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return BoostrapConfig{}, err
	}

	resp, err := sdk.sendRequest(req, key, string(CTJSON))
	if err != nil {
		return BoostrapConfig{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BoostrapConfig{}, err
	}

	if resp.StatusCode != http.StatusOK {
		if err := encodeError(resp.StatusCode); err != nil {
			return BoostrapConfig{}, err
		}
		return BoostrapConfig{}, ErrFetchFailed
	}

	var bc BoostrapConfig
	if err := json.Unmarshal(body, &bc); err != nil {
		return BoostrapConfig{}, err
	}

	return bc, nil
}

func (sdk mfSDK) UpdateBoostrap(key string, cfg BoostrapConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return ErrInvalidArgs
	}

	endpoint := fmt.Sprintf("%s/%s", "configs", cfg.MFThing)
	url := createURL(sdk.bootstrapURL, sdk.bootstrapPrefix, endpoint)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := sdk.sendRequest(req, key, string(CTJSON))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if err := encodeError(resp.StatusCode); err != nil {
			return err
		}
		return ErrFailedUpdate
	}

	return nil
}

func (sdk mfSDK) RemoveBoostrap(key, id string) error {
	endpoint := fmt.Sprintf("%s/%s", "configs", id)
	url := createURL(sdk.bootstrapURL, sdk.bootstrapPrefix, endpoint)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := sdk.sendRequest(req, key, string(CTJSON))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		if err := encodeError(resp.StatusCode); err != nil {
			return err
		}
		return ErrFailedRemoval
	}

	return nil
}

func (sdk mfSDK) Boostrap(key, id string) (BoostrapConfig, error) {
	endpoint := fmt.Sprintf("%s/%s", "bootstrap", id)
	url := createURL(sdk.bootstrapURL, sdk.bootstrapPrefix, endpoint)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return BoostrapConfig{}, err
	}

	resp, err := sdk.sendRequest(req, key, string(CTJSON))
	if err != nil {
		return BoostrapConfig{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BoostrapConfig{}, err
	}

	if resp.StatusCode != http.StatusOK {
		if err := encodeError(resp.StatusCode); err != nil {
			return BoostrapConfig{}, err
		}
		return BoostrapConfig{}, ErrFetchFailed
	}

	var bc BoostrapConfig
	if err := json.Unmarshal(body, &bc); err != nil {
		return BoostrapConfig{}, err
	}

	return bc, nil
}
