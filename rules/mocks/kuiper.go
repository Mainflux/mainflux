package mocks

import (
	"net/http"
	"strings"

	re "github.com/mainflux/mainflux/rules"
)

type kuiper struct {
	url     string
	streams map[string]re.Stream
	rules   map[string]re.Rule
}

var _ re.KuiperSDK = (*kuiper)(nil)

// NewKuiperSDK instantiates Kuiper API wrapper mock implementation.
func NewKuiperSDK(url string) re.KuiperSDK {
	return &kuiper{
		url:     url,
		streams: make(map[string]re.Stream),
		rules:   make(map[string]re.Rule),
	}
}

func (k *kuiper) Info() (re.Info, error) {
	return re.Info{
		Version:       "1.00",
		Os:            "Linux",
		UpTimeSeconds: 0,
	}, nil
}

func (k *kuiper) CreateStream(stream re.Stream) (*http.Response, error) {
	var res http.Response
	res.StatusCode = http.StatusCreated
	res.Body = http.NoBody

	if _, ok := k.streams[stream.Name]; ok {
		res.StatusCode = http.StatusConflict
		return &res, nil
	}

	k.streams[stream.Name] = stream

	return &res, nil
}

func (k *kuiper) UpdateStream(stream re.Stream) (*http.Response, error) {
	var res http.Response
	res.StatusCode = http.StatusNotFound
	res.Body = http.NoBody

	if _, ok := k.streams[stream.Name]; !ok {
		return &res, nil
	}

	k.streams[stream.Name] = stream
	res.StatusCode = http.StatusOK

	return &res, nil
}

func (k *kuiper) ShowStreams() ([]string, error) {
	var names []string

	for chanID := range k.streams {
		names = append(names, chanID)
	}

	return names, nil
}

func (k *kuiper) DescribeStream(name string) (*re.StreamInfo, error) {
	if _, ok := k.streams[name]; !ok {
		return &re.StreamInfo{}, re.ErrNotFound
	}

	info := re.StreamInfo{}
	fields(k.streams[name].Row, &info)
	info.Name = name

	return &info, nil
}

func (k *kuiper) Drop(name, kuiperType string) (*http.Response, error) {
	if _, ok := re.KuiperType[kuiperType]; !ok {
		return nil, re.ErrMalformedEntity
	}

	var res http.Response
	res.StatusCode = http.StatusBadRequest
	res.Body = http.NoBody

	if kuiperType == "streams" {
		if _, ok := k.streams[name]; !ok {
			return &res, nil
		}
		delete(k.streams, name)
		res.StatusCode = http.StatusOK
		return &res, nil
	}
	if kuiperType == "rules" {
		if _, ok := k.rules[name]; !ok {
			return &res, nil
		}
		delete(k.rules, name)
		res.StatusCode = http.StatusOK
		return &res, nil
	}

	return &res, nil
}

func (k *kuiper) CreateRule(rule re.Rule) (*http.Response, error) {
	var res http.Response
	res.StatusCode = http.StatusCreated
	res.Body = http.NoBody

	if _, ok := k.rules[rule.ID]; ok {
		res.StatusCode = http.StatusConflict
		return &res, nil
	}

	k.rules[rule.ID] = rule

	return &res, nil

}

func (k *kuiper) UpdateRule(rule re.Rule) (*http.Response, error) {
	var res http.Response
	res.StatusCode = http.StatusOK
	res.Body = http.NoBody

	if _, ok := k.rules[rule.ID]; !ok {
		res.StatusCode = http.StatusNotFound
		return &res, nil
	}

	k.rules[rule.ID] = rule

	return &res, nil
}

func (k *kuiper) ShowRules() ([]re.RuleInfo, error) {
	var ruleInfos []re.RuleInfo

	for _, v := range k.rules {
		ruleInfos = append(ruleInfos, re.RuleInfo{
			ID:     v.ID,
			Status: "Running",
		})
	}

	return ruleInfos, nil
}

func (k *kuiper) DescribeRule(name string) (*re.Rule, error) {
	if _, ok := k.rules[name]; !ok {
		return &re.Rule{}, re.ErrNotFound
	}

	r := k.rules[name]
	return &r, nil
}

func (k *kuiper) RuleStatus(name string) (map[string]interface{}, error) {
	var status map[string]interface{}
	if _, ok := k.rules[name]; !ok {
		return status, re.ErrNotFound
	}

	return status, nil
}

func (k *kuiper) ControlRule(name, action string) (*http.Response, error) {
	if _, ok := re.RuleAction[action]; !ok {
		return nil, re.ErrMalformedEntity
	}

	var res http.Response
	res.StatusCode = http.StatusOK
	res.Body = http.NoBody

	if _, ok := k.rules[name]; !ok {
		res.StatusCode = http.StatusNotFound
	}

	return &res, nil
}

func fields(sql string, info *re.StreamInfo) {
	if sql == "" {
		return
	}
	fieldArr := strings.Split(sql, ", ")
	for _, v := range fieldArr {
		info.StreamFields = append(info.StreamFields, struct {
			Name      string "json:\"Name\""
			FieldType string "json:\"FieldType\""
		}{strings.Split(v, " ")[0], strings.Split(v, " ")[1]})
	}
}
