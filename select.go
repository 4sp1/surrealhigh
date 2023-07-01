package surrealhigh

import (
	"errors"
	"fmt"
	"strings"

	"github.com/surrealdb/surrealdb.go"
)

type SurrealDB interface {
	Query(sql string, vars interface{}) (interface{}, error)
	Update(what string, data interface{}) (interface{}, error)
	Create(thing string, data interface{}) (interface{}, error)
}

type SurrealDriver interface {
	Driver() SurrealDB
}

type defaultDriver struct {
	db *surrealdb.DB
}

func (driver defaultDriver) Driver() SurrealDB {
	return driver.db
}

func DefaultDriver(db *surrealdb.DB) SurrealDriver {
	return defaultDriver{db}
}

type DBSelect[D Doc] interface {
	// Do returns with the following errors; in chronological order:
	// - type ErrDuplicateValuation
	// - any error from surrealdb.go query driver
	// - any error from surrealdb.go unmarshal
	// - ErrNoResult
	Do() ([]D, error)
}

type DBSelectAndUpdate[D Doc] interface {
	// Do returns with the following errors
	// - Any error from DBSelect.Do
	// - ErrNoDoc
	Do() (D, error)
}

func SelectOn[D Doc](q Select, db SurrealDriver) DBSelect[D] {
	return DBSelect[D](dbSelect[D]{
		query: q,
		db:    db,
	})
}

func SelectAndUpdate[D DocWithID](q Select, update func(D) D, db SurrealDriver) DBSelectAndUpdate[D] {
	return DBSelectAndUpdate[D](dbSelectUpdate[D]{
		query:  q,
		update: update,
		db:     db,
	})
}

type dbSelect[D Doc] struct {
	query Select
	db    SurrealDriver
}

type DocWithID interface {
	Doc
	Id() Thing
}

type dbSelectUpdate[D DocWithID] struct {
	query  Select
	update func(D) D
	db     SurrealDriver
}

type ErrDuplicateValuation struct {
	vars []conditionAtomVar
}

func (err ErrDuplicateValuation) Error() string {
	var couples []string
	for _, v := range err.vars {
		couples = append(couples, fmt.Sprintf("%s=%v", v.name, v.value))
	}
	return "found duplicate valuated condition atoms: " + strings.Join(couples, ", ")

}

var (
	ErrNoResult = errors.New("surrealdb: unmarshal results: no `results`")
)

func (q dbSelect[D]) Do() ([]D, error) {

	var vars map[string]interface{}

	if q.query.where != nil {

		vars = make(map[string]interface{})

		var duplicates []conditionAtomVar

		for _, v := range q.query.where.valuedVars() {
			if _, ok := vars[v.name.Var()]; ok {
				duplicates = append(duplicates, v)
			}
			vars[v.name.Var()] = v.value
		}

		if len(duplicates) > 0 {
			return nil, ErrDuplicateValuation{duplicates}
		}

	}

	data, err := q.db.Driver().Query(q.query.String(), vars)
	if err != nil {
		return nil, fmt.Errorf("surrealdb: %w", err)
	}

	var results []struct {
		Results []D    `json:"result"`
		Status  string `json:"status"`
	}

	// TODO see if there is a way to use SmartUnmarshal instead

	if err := surrealdb.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("surrealdb: unmarshal results: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrNoResult
	}

	if len(results[0].Results) == 0 {
		return nil, ErrNoResult
	}

	return results[0].Results, nil

}

var ErrNoDoc = errors.New("no document matched")

func (u dbSelectUpdate[D]) Do() (D, error) {
	q, db, update := u.query, u.db, u.update
	docs, err := SelectOn[D](q, db).Do()
	if err != nil {
		var d D
		return d, fmt.Errorf("select on %q: %w", d.Table(), err)
	}
	if len(docs) == 0 {
		var d D
		return d, fmt.Errorf("select on %q: %w", d.Table(), ErrNoDoc)
	}
	newDoc := update(docs[0])
	if _, err := db.Driver().Update(docs[0].Id().String(), newDoc); err != nil {
		return newDoc, fmt.Errorf("sdb: update %q: %w", docs[0].Id(), err)
	}
	return newDoc, nil
}
