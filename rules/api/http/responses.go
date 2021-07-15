// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/rules"
)

var _ mainflux.Response = (*infoRes)(nil)
var _ mainflux.Response = (*resultRes)(nil)
var _ mainflux.Response = (*listStreamsRes)(nil)
var _ mainflux.Response = (*viewStreamRes)(nil)
var _ mainflux.Response = (*listRulesRes)(nil)
var _ mainflux.Response = (*statusRes)(nil)

type infoRes struct {
	Version       string `json:"version"`
	Os            string `json:"os"`
	UpTimeSeconds int    `json:"upTimeSeconds"`
}

func (res infoRes) Code() int {
	return http.StatusOK
}

func (res infoRes) Headers() map[string]string {
	return map[string]string{}
}

func (res infoRes) Empty() bool {
	return false
}

type createRes struct {
	Result string `json:"result"`
}

func (res createRes) Code() int {
	return http.StatusCreated
}

func (res createRes) Headers() map[string]string {
	return map[string]string{}
}

func (res createRes) Empty() bool {
	return false
}

type resultRes struct {
	Result string `json:"result"`
}

func (res resultRes) Code() int {
	return http.StatusOK
}

func (res resultRes) Headers() map[string]string {
	return map[string]string{}
}

func (res resultRes) Empty() bool {
	return false
}

type listStreamsRes struct {
	Streams []string `json:"streams"`
}

func (res listStreamsRes) Code() int {
	return http.StatusOK
}

func (res listStreamsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listStreamsRes) Empty() bool {
	return false
}

type viewStreamRes struct {
	Stream rules.StreamInfo
}

func (res viewStreamRes) Code() int {
	return http.StatusOK
}

func (res viewStreamRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewStreamRes) Empty() bool {
	return false
}

type listRulesRes struct {
	Rules []rules.RuleInfo
}

func (res listRulesRes) Code() int {
	return http.StatusOK
}

func (res listRulesRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRulesRes) Empty() bool {
	return false
}

type viewRuleRes struct {
	Rule rules.Rule
}

func (res viewRuleRes) Code() int {
	return http.StatusOK
}

func (res viewRuleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewRuleRes) Empty() bool {
	return false
}

type statusRes map[string]interface{}

func (res statusRes) Code() int {
	return http.StatusOK
}

func (res statusRes) Headers() map[string]string {
	return map[string]string{}
}

func (res statusRes) Empty() bool {
	return false
}

type errorRes struct {
	Err string `json:"error"`
}
