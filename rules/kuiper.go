package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	// RuleAction is a map of sink actions
	RuleAction = map[string]string{"start": "start", "stop": "stop", "restart": "restart"}
	// KuiperType is a map of kuiper base entities
	KuiperType = map[string]string{"streams": "streams", "rules": "rules"}
)

// KuiperSDK specifies an API that must be fullfiled by KuiperSDK implementations
type KuiperSDK interface {
	// Info gets the version number, system type, and Kuiper running time
	Info() (Info, error)

	// CreateStream creates a Kuiper stream
	CreateStream(stream Stream) (*http.Response, error)

	// UpdateStream updates a stream definition
	UpdateStream(stream Stream) (*http.Response, error)

	// ShowStreams displays all defined streams
	ShowStreams() ([]string, error)

	// DescribeStream prints the detailed definition of a stream
	DescribeStream(name string) (*StreamInfo, error)

	// Drop deletes the stream or rule definition
	Drop(name, kuiperType string) (*http.Response, error)

	// CreateRule creates and starts a rule
	CreateRule(rule Rule) (*http.Response, error)

	// UpdateRule updates a rule
	UpdateRule(rule Rule) (*http.Response, error)

	// ShowRules displays all of rules with a brief status
	ShowRules() ([]RuleInfo, error)

	// DescribeRule prints the detailed rule definition
	DescribeRule(name string) (*Rule, error)

	// RuleStatus gets the rule status
	RuleStatus(name string) (map[string]interface{}, error)

	// ControlRule starts, stops or restarts the rule
	ControlRule(name, action string) (*http.Response, error)
}

type kuiper struct {
	url        string
	pluginHost string
	pluginPort string
}

var _ KuiperSDK = (*kuiper)(nil)

// NewKuiperSDK instantiates KuiperSDK
func NewKuiperSDK(url, pluginHost, pluginPort string) KuiperSDK {
	return &kuiper{
		url:        url,
		pluginHost: pluginHost,
		pluginPort: pluginPort,
	}
}

func (k *kuiper) Info() (Info, error) {
	var i Info

	res, err := http.Get(k.url)
	if err != nil {
		return i, errors.Wrap(ErrKuiperServer, err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&i)
	if err != nil {
		return i, errors.Wrap(ErrKuiperServer, err)
	}

	return i, nil
}

func (k *kuiper) CreateStream(stream Stream) (*http.Response, error) {
	sql := k.sql(&stream)
	body, err := json.Marshal(map[string]string{"sql": sql})
	if err != nil {
		return nil, ErrMalformedEntity
	}

	url := fmt.Sprintf("%s/streams", k.url)
	res, err := http.Post(url, contentType, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}

	return res, nil
}

func (k *kuiper) UpdateStream(stream Stream) (*http.Response, error) {
	sql := k.sql(&stream)
	body, err := json.Marshal(map[string]string{"sql": sql})
	if err != nil {
		return nil, ErrMalformedEntity
	}

	url := fmt.Sprintf("%s/streams/%s", k.url, stream.Name)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, ErrMalformedEntity
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}

	return res, nil
}

func (k *kuiper) ShowStreams() ([]string, error) {
	var streams []string

	url := fmt.Sprintf("%s/streams", k.url)
	res, err := http.Get(url)
	if err != nil {
		return streams, errors.Wrap(ErrKuiperServer, err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&streams)
	if err != nil {
		return streams, errors.Wrap(ErrMalformedEntity, err)
	}
	return streams, nil
}

func (k *kuiper) DescribeStream(name string) (*StreamInfo, error) {
	url := fmt.Sprintf("%s/streams/%s", k.url, name)
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, errors.Wrap(ErrNotFound, err)
	}
	defer res.Body.Close()

	var stream StreamInfo
	err = json.NewDecoder(res.Body).Decode(&stream)
	if err != nil {
		return nil, errors.Wrap(ErrMalformedEntity, err)
	}

	return &stream, nil
}

func (k *kuiper) Drop(name, kuiperType string) (*http.Response, error) {
	if _, ok := KuiperType[kuiperType]; !ok {
		return nil, ErrMalformedEntity
	}

	url := fmt.Sprintf("%s/%s/%s", k.url, kuiperType, name)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}
	return res, nil
}

func (k *kuiper) CreateRule(rule Rule) (*http.Response, error) {
	k.ruleURL(&rule)

	body, err := json.Marshal(rule)
	if err != nil {
		return nil, ErrMalformedEntity
	}

	url := fmt.Sprintf("%s/rules", k.url)
	res, err := http.Post(url, contentType, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}
	return res, nil
}

func (k *kuiper) UpdateRule(rule Rule) (*http.Response, error) {
	k.ruleURL(&rule)

	body, err := json.Marshal(rule)
	if err != nil {
		return nil, ErrMalformedEntity
	}

	url := fmt.Sprintf("%s/rules/%s", k.url, rule.ID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}

	return res, nil
}

func (k *kuiper) ShowRules() ([]RuleInfo, error) {
	var ruleInfos []RuleInfo

	url := fmt.Sprintf("%s/rules", k.url)
	res, err := http.Get(url)
	if err != nil {
		return ruleInfos, errors.Wrap(ErrKuiperServer, err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&ruleInfos)
	if err != nil {
		return ruleInfos, errors.Wrap(ErrMalformedEntity, err)
	}

	return ruleInfos, nil
}

func (k *kuiper) DescribeRule(name string) (*Rule, error) {
	url := fmt.Sprintf("%s/rules/%s", k.url, name)

	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, errors.Wrap(ErrNotFound, err)
	}
	defer res.Body.Close()

	var rule Rule
	err = json.NewDecoder(res.Body).Decode(&rule)
	if err != nil {
		return nil, errors.Wrap(ErrMalformedEntity, err)
	}

	return &rule, nil
}

func (k *kuiper) RuleStatus(name string) (map[string]interface{}, error) {
	var status map[string]interface{}

	url := fmt.Sprintf("%s/rules/%s/status", k.url, name)

	res, err := http.Get(url)
	if err != nil {
		return status, errors.Wrap(ErrKuiperServer, err)
	}
	if res.StatusCode == http.StatusNotFound {
		return status, errors.Wrap(ErrNotFound, err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&status)
	if err != nil {
		return status, errors.Wrap(ErrMalformedEntity, err)
	}

	return status, nil
}

func (k *kuiper) ControlRule(name, action string) (*http.Response, error) {
	if _, ok := RuleAction[action]; !ok {
		return nil, ErrMalformedEntity
	}

	url := fmt.Sprintf("%s/rules/%s/%s", k.url, name, action)
	res, err := http.Post(url, "", nil)
	if err != nil {
		return nil, errors.Wrap(ErrKuiperServer, err)
	}
	return res, nil
}

func (k *kuiper) sql(stream *Stream) string {
	if stream.Host == "" {
		stream.Host = k.pluginHost
	}
	if stream.Port == "" {
		stream.Port = k.pluginPort
	}
	url := fmt.Sprintf("%s:%s", stream.Host, stream.Port)
	// Kuiper source will unpack topic in url and topic
	topic := fmt.Sprintf("%s;%s", url, stream.Channel)
	if len(stream.Subtopic) > 0 {
		topic = fmt.Sprintf("%s.%s", topic, stream.Subtopic)
	}
	return fmt.Sprintf("create stream %s (%s) WITH (DATASOURCE = \"%s\" FORMAT = \"%s\" TYPE = \"%s\")", stream.Name, stream.Row, topic, format, pluginType)
}

func (k *kuiper) ruleURL(rule *Rule) {
	if rule.Actions[0].Mainflux.Host == "" {
		rule.Actions[0].Mainflux.Host = k.pluginHost
	}
	if rule.Actions[0].Mainflux.Port == "" {
		rule.Actions[0].Mainflux.Port = k.pluginPort
	}
}
