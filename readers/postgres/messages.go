// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx" // required for DB access
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/pkg/errors"
	jsont "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
)

const errInvalid = "invalid_text_representation"

const (
	format = "format"
	// Table for SenML messages
	defTable = "messages"

	// Error code for Undefined table error.
	undefinedTableCode = "42P01"
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

func (tr postgresRepository) ReadAll(pm readers.PageMetadata) (readers.MessagesPage, error) {
	order := "time"
	table := defTable

	ft, ok := pm.Query[format].(string)
	if !ok {
		return readers.MessagesPage{}, errReadMessages
	}
	if ft != "" && ft != defTable {
		order = "created"
		table = ft
	}

	params := pm.Query
	delete(params, format)
	fmt.Println("PARAMS:", params)
	condition := fmtCondition(params)

	q := fmt.Sprintf(`SELECT * FROM %s
    WHERE %s ORDER BY %s DESC
	LIMIT :limit OFFSET :offset;`, table, condition, order)
	params["limit"] = int(pm.Limit)
	params["offset"] = int(pm.Offset)

	// params := map[string]interface{}{
	// 	"channel":      chanID,
	// 	"limit":        rpm.Limit,
	// 	"offset":       rpm.Offset,
	// 	"subtopic":     rpm.Subtopic,
	// 	"publisher":    rpm.Publisher,
	// 	"name":         rpm.Name,
	// 	"protocol":     rpm.Protocol,
	// 	"value":        rpm.Value,
	// 	"bool_value":   rpm.BoolValue,
	// 	"string_value": rpm.StringValue,
	// 	"data_value":   rpm.DataValue,
	// 	"from":         rpm.From,
	// 	"to":           rpm.To,
	// }
	fmt.Println("QUERY:", q, "\nPARMAS:", params, "\nCONDITION:", condition)

	rows, err := tr.db.NamedQuery(q, params)
	if err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == undefinedTableCode {
				return readers.MessagesPage{}, nil
			}
		}
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	defer rows.Close()

	page := readers.MessagesPage{
		PageMetadata: pm,
		Messages:     []readers.Message{},
	}
	switch format {
	case defTable:
		for rows.Next() {
			msg := senmlMessage{Message: senml.Message{}}
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
			m["payload"] = jsont.ParseFlat(m["payload"])
			page.Messages = append(page.Messages, m)
		}

	}

	q = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s;`, format, condition)
	rows, err = tr.db.NamedQuery(q, params)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	defer rows.Close()

	total := uint64(0)
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return page, err
		}
	}
	page.Total = total

	return page, nil
}

func fmtCondition(query map[string]interface{}) string {
	condition := `channel = :channel`
	// var condition string
	// delete(query, "channel")
	conds := []string{}
	for name := range query {
		switch name {
		case "value":
			comparator := readers.ParseValueComparator(query)
			condition = fmt.Sprintf(`%s AND value %s :value`, condition, comparator)
			conds = append(conds, fmt.Sprintf("value %s :value", comparator))
		// case "bool_value":
		// 	condition = fmt.Sprintf(`%s AND bool_value = :bool_value`, condition)
		// case "unot":
		// 	condition = fmt.Sprintf(`%s AND bool_value = :bool_value`, condition)
		// case "string_value":
		// 	condition = fmt.Sprintf(`%s AND string_value = :string_value`, condition)
		// case "data_value":
		// 	condition = fmt.Sprintf(`%s AND data_value = :data_value`, condition)
		case "from":
			condition = fmt.Sprintf(`%s AND time >= :from`, condition)
			conds = append(conds, "time >= :from")
		case "to":
			condition = fmt.Sprintf(`%s AND time < :to`, condition)
			conds = append(conds, "time < :to")
		default:
			condition = fmt.Sprintf(`%s AND %s = :%s`, condition, name, name)
			conds = append(conds, fmt.Sprintf("%s = :%s", name, name))
		}
	}
	// fmt.Println(strings.Join(conds, " AND "))
	// fmt.Println("CNDTN:", condition)
	// fmt.Println("QUERY:", query)
	return strings.Join(conds, " AND ")
}

type senmlMessage struct {
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
