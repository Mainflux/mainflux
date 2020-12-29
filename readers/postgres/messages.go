// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx" // required for DB access
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
)

const errInvalid = "invalid_text_representation"

const (
	format   = "format"
	defTable = "messages"
)

var errReadMessages = errors.New("failed to read messages from postgres database")

var _ readers.MessageRepository = (*postgresRepository)(nil)

type postgresRepository struct {
	db *sqlx.DB
}

// New returns new PostgreSQL writer.
func New(db *sqlx.DB) readers.MessageRepository {
	return &postgresRepository{
		db: db,
	}
}

func (tr postgresRepository) ReadAll(chanID string, offset, limit uint64, query map[string]string) (readers.MessagesPage, error) {
	table, ok := query[format]
	order := "created"
	if !ok {
		table = defTable
		order = "time"
	}
	// Remove format filter and format the rest properly.
	delete(query, format)
	q := fmt.Sprintf(`SELECT * FROM %s
    WHERE %s ORDER BY %s DESC
	LIMIT :limit OFFSET :offset;`, table, fmtCondition(chanID, query), order)
	fmt.Println("QUERY", q)

	params := map[string]interface{}{
		"channel":   chanID,
		"limit":     limit,
		"offset":    offset,
		"subtopic":  query["subtopic"],
		"publisher": query["publisher"],
		"name":      query["name"],
		"protocol":  query["protocol"],
	}

	rows, err := tr.db.NamedQuery(q, params)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	defer rows.Close()

	page := readers.MessagesPage{
		Offset:   offset,
		Limit:    limit,
		Messages: []interface{}{},
	}
	switch table {
	case defTable:
		for rows.Next() {
			msg := dbMessage{Message: senml.Message{Channel: chanID}}
			if err := rows.StructScan(&msg); err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}

			page.Messages = append(page.Messages, msg.Message)
		}
	default:
		for rows.Next() {
			msg := jsonMessage{}
			if err := rows.StructScan(&msg); err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}
			m, err := msg.toMap()
			if err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}
			m["payload"] = parseFlat(m["payload"])
			page.Messages = append(page.Messages, parseFlat(m))
		}

	}

	q = `SELECT COUNT(*) FROM messages WHERE channel = $1;`
	qParams := []interface{}{chanID}

	if query["subtopic"] != "" {
		q = `SELECT COUNT(*) FROM messages WHERE channel = $1 AND subtopic = $2;`
		qParams = append(qParams, query["subtopic"])
	}

	if err := tr.db.QueryRow(q, qParams...).Scan(&page.Total); err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}

	return page, nil
}

func fmtCondition(chanID string, query map[string]string) string {
	condition := `channel = :channel`
	for name := range query {
		switch name {
		case
			"subtopic",
			"publisher",
			"name",
			"protocol":
			condition = fmt.Sprintf(`%s AND %s = :%s`, condition, name, name)
		}
	}
	return condition
}

func parseFlat(flat interface{}) interface{} {
	msg := make(map[string]interface{})
	switch v := flat.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if value == nil {
				continue
			}
			keys := strings.Split(key, "/")
			n := len(keys)
			if n == 1 {
				msg[key] = value
				continue
			}
			current := msg
			for i, k := range keys {
				if _, ok := current[k]; !ok {
					current[k] = make(map[string]interface{})
				}
				if i == n-1 {
					current[k] = value
					break
				}
				current = current[k].(map[string]interface{})
			}
		}
	}
	return msg
}

type dbMessage struct {
	ID string `db:"id"`
	senml.Message
}

type jsonMessage struct {
	ID        string `db:"id"`
	Channel   string `db:"channel"`
	Created   int64  `db:"created"`
	Subtopic  string `db:"subtopic"`
	Publisher string `db:"publisher"`
	Protocol  string `db:"protocol"`
	Payload   []byte `db:"payload"`
}

func (msg jsonMessage) toMap() (map[string]interface{}, error) {
	ret := map[string]interface{}{
		"id":        msg.ID,
		"channel":   msg.Channel,
		"created":   msg.Created,
		"subtopic":  msg.Subtopic,
		"publisher": msg.Publisher,
		"protocol":  msg.Protocol,
		"payload":   map[string]interface{}{},
	}
	pld := make(map[string]interface{})
	if err := json.Unmarshal(msg.Payload, &pld); err != nil {
		return nil, err
	}
	ret["payload"] = pld
	return ret, nil
}
