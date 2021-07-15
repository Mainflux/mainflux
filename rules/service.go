// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rules

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
)

const contentType = "application/json"

var (
	prefixStart = "id_"
	prefixEnd   = "_"
	remIDRegEx  = regexp.MustCompile(prefixStart + "[[:alnum:]]{32}" + prefixEnd)

	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrKuiperServer indicates internal kuiper rules engine server error
	ErrKuiperServer = errors.New("kuiper error")
)

// Info is used to fetch kuiper running instance data
type Info struct {
	Version       string `json:"version"`
	Os            string `json:"os"`
	UpTimeSeconds int    `json:"upTimeSeconds"`
}

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Info fetches kuiper version, os and uptime information
	Info(ctx context.Context) (Info, error)

	// CreateStream creates kuiper stream
	CreateStream(ctx context.Context, token string, stream Stream) (string, error)

	// UpdateStream updates kuiper stream
	UpdateStream(ctx context.Context, token string, stream Stream) (string, error)

	// ListStream list kuiper streams
	ListStreams(ctx context.Context, token string) ([]string, error)

	// ViewStream fetches kuiper stream info
	ViewStream(ctx context.Context, token, name string) (StreamInfo, error)

	// Delete deletes streams and rules depending on the value of kuiperType parameter
	Delete(ctx context.Context, token, name, kuiperType string) (string, error)

	// CreateRule creates kuiper rule
	CreateRule(ctx context.Context, token string, rule Rule) (string, error)

	// UpdateRule updates kuiper rule
	UpdateRule(ctx context.Context, token string, rule Rule) (string, error)

	// ListRules list kuiper rules
	ListRules(ctx context.Context, token string) ([]RuleInfo, error)

	// ViewRule fetches kuiper rule info
	ViewRule(ctx context.Context, token, name string) (Rule, error)

	// RuleStatus fetches kuiper rule operation info
	RuleStatus(ctx context.Context, token, name string) (map[string]interface{}, error)

	// ControlRule is used to "start", "stop" and "restart" kuiper rule
	ControlRule(ctx context.Context, token, name, action string) (string, error)
}

var actions = map[string]string{
	"createStream": "Create stream",
	"updateStream": "Update stream",
	"createRule":   "Create rule",
	"updateRule":   "Update rule",
	"delete":       "Delete",
}

var results = map[string]string{
	"success": "successful",
	"fail":    "failed",
}

type reService struct {
	auth   mainflux.AuthServiceClient
	things mainflux.ThingsServiceClient
	kuiper KuiperSDK
	logger logger.Logger
}

var _ Service = (*reService)(nil)

// New instantiates the rules service implementation.
func New(kuiper KuiperSDK, auth mainflux.AuthServiceClient, things mainflux.ThingsServiceClient, logger logger.Logger) Service {
	return &reService{
		auth:   auth,
		things: things,
		kuiper: kuiper,
		logger: logger,
	}
}

func (re *reService) Info(_ context.Context) (Info, error) {
	return re.kuiper.Info()
}

func (re *reService) CreateStream(ctx context.Context, token string, stream Stream) (string, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	if _, err = re.things.IsChannelOwner(ctx, &mainflux.ChannelOwnerReq{
		Owner:  ui.Email,
		ChanID: stream.Channel,
	}); err != nil {
		return "", ErrNotFound
	}

	stream.Name = prepend(ui.Id, stream.Name)
	res, err := re.kuiper.CreateStream(stream)
	if err != nil {
		return "", err
	}

	result, err := makeResult(res, actions["createStream"], http.StatusCreated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (re *reService) UpdateStream(ctx context.Context, token string, stream Stream) (string, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	if _, err = re.things.IsChannelOwner(ctx, &mainflux.ChannelOwnerReq{
		Owner:  ui.Email,
		ChanID: stream.Channel,
	}); err != nil {
		return "", ErrNotFound
	}

	stream.Name = prepend(ui.Id, stream.Name)
	res, err := re.kuiper.UpdateStream(stream)
	if err != nil {
		return "", err
	}

	result, err := makeResult(res, actions["updateStream"], http.StatusOK)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (re *reService) ListStreams(ctx context.Context, token string) ([]string, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return []string{}, ErrUnauthorizedAccess
	}

	streams, err := re.kuiper.ShowStreams()
	if err != nil {
		return streams, err
	}

	var owned []string
	for _, value := range streams {
		if !strings.HasPrefix(value, prefix(ui.Id)) {
			continue
		}
		owned = append(owned, remove(ui.Id, value))
	}

	return owned, nil
}

func (re *reService) ViewStream(ctx context.Context, token, name string) (StreamInfo, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return StreamInfo{}, ErrUnauthorizedAccess
	}

	stream, err := re.kuiper.DescribeStream(prepend(ui.Id, name))
	if err != nil {
		return StreamInfo{}, err
	}

	stream.Name = remove(ui.Id, stream.Name)

	return *stream, nil
}

func (re *reService) Delete(ctx context.Context, token, name, kuiperType string) (string, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	res, err := re.kuiper.Drop(prepend(ui.Id, name), kuiperType)
	if err != nil {
		return "", err
	}

	result, err := makeResult(res, fmt.Sprintf("%s %s", actions["Delete"], kuiperType), http.StatusOK)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (re *reService) CreateRule(ctx context.Context, token string, rule Rule) (string, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	if _, err = re.things.IsChannelOwner(ctx, &mainflux.ChannelOwnerReq{
		Owner:  ui.Email,
		ChanID: rule.Actions[0].Mainflux.Channel,
	}); err != nil {
		return "", ErrNotFound
	}

	rulePrependInplace(ui.Id, &rule)
	res, err := re.kuiper.CreateRule(rule)
	if err != nil {
		return "", err
	}

	result, err := makeResult(res, actions["createRule"], http.StatusCreated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (re *reService) UpdateRule(ctx context.Context, token string, rule Rule) (string, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	if _, err = re.things.IsChannelOwner(ctx, &mainflux.ChannelOwnerReq{
		Owner:  ui.Email,
		ChanID: rule.Actions[0].Mainflux.Channel,
	}); err != nil {
		return "", ErrNotFound
	}

	rulePrependInplace(ui.Id, &rule)
	res, err := re.kuiper.UpdateRule(rule)
	if err != nil {
		return "", err
	}

	result, err := makeResult(res, actions["updateRule"], http.StatusOK)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (re *reService) ListRules(ctx context.Context, token string) ([]RuleInfo, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return []RuleInfo{}, ErrUnauthorizedAccess
	}

	rules, err := re.kuiper.ShowRules()
	if err != nil {
		return nil, err
	}

	var owned []RuleInfo
	for i, value := range rules {
		if !strings.HasPrefix(value.ID, prefix(ui.Id)) {
			continue
		}
		rules[i].ID = remove(ui.Id, value.ID)
		owned = append(owned, rules[i])
	}

	return owned, nil
}

func (re *reService) ViewRule(ctx context.Context, token, name string) (Rule, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Rule{}, ErrUnauthorizedAccess
	}

	rule, err := re.kuiper.DescribeRule(prepend(ui.Id, name))
	if err != nil {
		return Rule{}, err
	}

	ruleRemove(ui.Id, rule)

	return *rule, nil
}

func (re *reService) RuleStatus(ctx context.Context, token, name string) (map[string]interface{}, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return map[string]interface{}{}, ErrUnauthorizedAccess
	}

	status, err := re.kuiper.RuleStatus(prepend(ui.Id, name))
	if err != nil {
		return nil, err
	}

	for key, val := range status {
		delete(status, key)
		status[remove(ui.Id, key)] = val
	}

	return status, nil
}

func (re *reService) ControlRule(ctx context.Context, token, name, action string) (string, error) {
	ui, err := re.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	res, err := re.kuiper.ControlRule(prepend(ui.Id, name), action)
	if err != nil {
		return "", err
	}

	result, err := makeResult(res, action, http.StatusOK)
	if err != nil {
		return "", err
	}
	return result, nil
}

func prefix(id string) string {
	return prefixStart + strings.ReplaceAll(id, "-", "") + prefixEnd
}

func prepend(id, name string) string {
	return fmt.Sprintf("%s%s", prefix(id), name)
}

func remove(id, name string) string {
	return strings.Replace(name, prefix(id), "", 1)
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func rulePrependInplace(id string, rule *Rule) *Rule {
	rule.ID = prepend(id, rule.ID)

	// Prepend id to stream name; sql e.g. "select * from stream_name where v > 1.2;"
	words := strings.Fields(rule.SQL)
	idx := indexOf("from", words) + 1
	words[idx] = prepend(id, words[idx])
	rule.SQL = strings.Join(words, " ")

	return rule
}

func ruleRemove(id string, rule *Rule) {
	rule.ID = remove(id, rule.ID)

	words := strings.Fields(rule.SQL)
	idx := indexOf("from", words) + 1
	words[idx] = remove(id, words[idx])
	rule.SQL = strings.Join(words, " ")
}

func makeResult(res *http.Response, action string, status int) (string, error) {
	result := fmt.Sprintf("%s %s", action, results["success"])
	if res.StatusCode != status {
		defer res.Body.Close()
		reasonBt, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", errors.Wrap(ErrKuiperServer, err)
		}
		reasonStr := remIDRegEx.ReplaceAllString(string(reasonBt), "")
		result = fmt.Sprintf("%s %s. Kuiper http status %d: %s", action, results["fail"], res.StatusCode, reasonStr)
		return "", errors.Wrap(ErrKuiperServer, errors.New(result))
	}
	return result, nil
}
