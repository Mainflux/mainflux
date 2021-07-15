package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	FORMAT_BINARY = "binary"
	FORMAT_JSON   = "json"

	DEFAULT_FIELD = "self"
)

func MessageDecode(payload []byte, format string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	switch strings.ToLower(format) {
	case FORMAT_JSON:
		e := json.Unmarshal(payload, &result)
		return result, e
	case FORMAT_BINARY:
		result[DEFAULT_FIELD] = payload
		return result, nil
	}
	return nil, fmt.Errorf("invalid format %s", format)
}
