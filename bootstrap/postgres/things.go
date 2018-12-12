package postgres

import (
	"database/sql"
	"fmt"
	"nov/bootstrap"

	"github.com/lib/pq" // required for DB access
	"github.com/mainflux/mainflux/logger"
)

const duplicateErr = "unique_violation"

var _ bootstrap.ThingRepository = (*thingRepository)(nil)

type thingRepository struct {
	db  *sql.DB
	log logger.Logger
}

// NewThingRepository instantiates a PostgreSQL implementation of thing
// repository.
func NewThingRepository(db *sql.DB, log logger.Logger) bootstrap.ThingRepository {
	return &thingRepository{db: db, log: log}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (tr thingRepository) Save(thing bootstrap.Thing) (string, error) {
	q := `INSERT INTO things (mainflux_key, owner, mainflux_thing, external_id, external_key, mainflux_channels, config, state)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	if err := tr.db.QueryRow(q, thing.MFKey, thing.Owner, thing.MFThing, thing.ExternalID, thing.ExternalKey, pq.Array(thing.MFChannels),
		nullString(thing.Config), thing.State).Scan(&thing.ID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			return "", bootstrap.ErrConflict
		}
		return "", err
	}

	return thing.ID, nil
}

func (tr thingRepository) RetrieveByID(key, id string) (bootstrap.Thing, error) {
	q := `SELECT mainflux_key, mainflux_thing, external_id, external_key, mainflux_channels, config, state FROM things WHERE id = $1 AND owner = $2`
	thing := bootstrap.Thing{ID: id, Owner: key}
	var mfKey, mfThing, config sql.NullString

	err := tr.db.
		QueryRow(q, id, key).
		Scan(&mfKey, &mfThing, &thing.ExternalID, &thing.ExternalKey, pq.Array(&thing.MFChannels), &config, &thing.State)

	thing.MFKey = mfKey.String
	thing.MFThing = mfThing.String
	thing.Config = config.String

	if err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	return thing, nil
}

func (tr thingRepository) RetrieveAll(key string, state bootstrap.State, offset, limit uint64) []bootstrap.Thing {
	rows, err := tr.retrieveAll(key, state, offset, limit)
	if err != nil {
		tr.log.Error(fmt.Sprintf("Failed to retrieve things due to %s", err))
		return []bootstrap.Thing{}
	}
	defer rows.Close()

	items := []bootstrap.Thing{}
	var mfKey, mfThing, config sql.NullString
	for rows.Next() {
		t := bootstrap.Thing{Owner: key}
		if err = rows.Scan(&mfKey, &mfThing, &t.ExternalID, &t.ExternalKey, pq.Array(&t.MFChannels), &config, &t.State); err != nil {
			tr.log.Error(fmt.Sprintf("Failed to read retrieved thing due to %s", err))

			return []bootstrap.Thing{}
		}

		t.MFKey = mfKey.String
		t.MFThing = mfThing.String
		t.Config = config.String

		items = append(items, t)
	}

	return items
}

func (tr thingRepository) RetrieveByExternalID(externalKey, externalID string) (bootstrap.Thing, error) {
	q := `SELECT id, owner, mainflux_key, mainflux_thing, mainflux_channels, config, state FROM things WHERE external_key = $1 AND external_id = $2`

	var mfKey, mfThing, config sql.NullString
	thing := bootstrap.Thing{
		ExternalID:  externalID,
		ExternalKey: externalKey,
	}
	if err := tr.db.QueryRow(q, externalKey, externalID).Scan(&thing.ID, &thing.Owner, &mfKey, &mfThing, pq.Array(&thing.MFChannels), &config, &thing.State); err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	thing.MFKey = mfKey.String
	thing.MFThing = mfThing.String
	thing.Config = config.String

	return thing, nil
}

func (tr thingRepository) Update(thing bootstrap.Thing) error {
	q := `UPDATE things SET mainflux_channels = $1, config = $2, state = $3 WHERE id = $4 AND owner = $5`
	res, err := tr.db.Exec(q, pq.Array(thing.MFChannels), thing.Config, thing.State, thing.ID, thing.Owner)
	if err != nil {
		return err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if cnt == 0 {
		return bootstrap.ErrNotFound
	}

	return nil
}

func (tr thingRepository) Remove(key, id string) error {
	q := `DELETE FROM things WHERE id = $1 AND owner = $2`
	tr.db.Exec(q, id, key)
	return nil
}

func (tr thingRepository) ChangeState(key, id string, state bootstrap.State) error {
	q := `UPDATE things SET state = $1 WHERE id = $2 AND owner = $3;`

	res, err := tr.db.Exec(q, state, id, key)
	if err != nil {
		return err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if cnt == 0 {
		return bootstrap.ErrNotFound
	}

	return nil
}

func (tr thingRepository) retrieveAll(key string, state bootstrap.State, offset, limit uint64) (*sql.Rows, error) {
	template := `SELECT mainflux_key, mainflux_thing, external_id, external_key, mainflux_channels, config, state FROM things WHERE %s %s ORDER BY id LIMIT %s OFFSET %s`
	if key == "" {
		template = fmt.Sprintf(template, "owner = $1")
	}
	if state != -1 {
		q := fmt.Sprintf(template, "", "$2", "$3")
		return tr.db.Query(q, key, limit, offset)
	}

	q := fmt.Sprintf(template, "AND state = $2", "$3", "$4")
	return tr.db.Query(q, key, state, limit, offset)
}
