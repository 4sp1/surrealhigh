package surrealhigh

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/surrealdb/surrealdb.go"
)

type Doc interface {
	Table() Table
}

type DBDoc interface {
	Doc

	json.Marshaler

	DB() *surrealdb.DB
	Create() (Id, error)
}

func NewDefaultDoc(doc Doc, db *surrealdb.DB) DefaultDoc {
	return DefaultDoc{
		doc: doc,
		db:  db,
	}
}

type DefaultDoc struct {
	doc Doc
	db  *surrealdb.DB
}

var _ DBDoc = DefaultDoc{}

func (doc DefaultDoc) Table() Table {
	return doc.doc.Table()
}

func (doc DefaultDoc) MarshalJSON() ([]byte, error) {
	return json.Marshal(doc.doc)
}

func (doc DefaultDoc) DB() *surrealdb.DB {
	return doc.db
}

var nilID = Id(uuid.Nil)

func (doc DefaultDoc) Create() (Id, error) {
	errWrapf := func(f string, a ...interface{}) (Id, error) {
		return nilID, fmt.Errorf(f, a...)
	}
	data, err := doc.db.Create(string(NewID().Thing(doc.Table())), doc.doc)
	if err != nil {
		return errWrapf("sdb: create: %w", err)
	}
	var docId struct {
		Id string `json:"id"`
	}
	if err := surrealdb.Unmarshal(data, &docId); err != nil {
		return errWrapf("unmarshal docId: %w", err)
	}
	recId, err := parseId(docId.Id)
	if err != nil {
		return errWrapf("hsdb: parse id: %w", err)
	}
	return recId, nil
}

func parseId(id string) (Id, error) {
	s := strings.Split(id[:len(id)-2], "⟨")
	if len(s) != 2 {
		return nilID, fmt.Errorf("unknown id format %q", id)
	}
	for i, r := range s[1] {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' {
			u, err := uuid.Parse(s[1][0:i])
			if err != nil {
				return nilID, fmt.Errorf("uuid: parse: %w", err)
			}
			return Id(u), nil
		}
	}
	return nilID, fmt.Errorf("unknown id format %q", id)
}