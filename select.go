package surrealhigh

import (
	"errors"
	"fmt"
	"strings"

	"github.com/surrealdb/surrealdb.go"
)

type surrealDriver interface {
	Query(sql string, vars interface{}) (interface{}, error)
}

type SelectDriver interface {
	driver() surrealDriver
}

type defaultDriver struct {
	db *surrealdb.DB
}

func (driver defaultDriver) driver() surrealDriver {
	return driver.db
}

func DefaultDriver(db *surrealdb.DB) SelectDriver {
	return defaultDriver{db}
}

type DBSelect[D Doc] interface {
	Do() ([]D, error)
}

func SelectOn[D Doc](q Select, db SelectDriver) DBSelect[D] {
	return DBSelect[D](dbSelect[D]{
		query: q,
		db:    db,
	})
}

type dbSelect[D Doc] struct {
	query Select
	db    SelectDriver
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

	data, err := q.db.driver().Query(q.query.String(), vars)
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

	return results[0].Results, nil

}
