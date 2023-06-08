package surrealhigh

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/surrealdb/surrealdb.go"
)

type Doc interface {
	Table() Table
}

type doc struct {
	from Table
}

func (doc doc) Table() Table {
	return doc.from
}

// DBDoc is a low level document as it has a reference to the database driver
type DBDoc interface {
	Doc
	json.Marshaler
	Create() (Id, error)

	db() surrealDriver
}

func NewDefaultDoc(doc Doc, db SurrealDriver) DefaultDoc {
	return DefaultDoc{
		doc:    doc,
		driver: db.driver(),
	}
}

type DefaultDoc struct {
	doc    Doc
	driver surrealDriver
}

var _ DBDoc = DefaultDoc{}

func (doc DefaultDoc) Table() Table {
	return doc.doc.Table()
}

func (doc DefaultDoc) MarshalJSON() ([]byte, error) {
	return json.Marshal(doc.doc)
}

func (doc DefaultDoc) db() surrealDriver {
	return doc.driver
}

var nilID = Id(uuid.Nil)

func (doc DefaultDoc) Create() (Id, error) {

	// TODO(malikbenkirane) rm
	errWrapf := func(f string, a ...interface{}) (Id, error) {
		return nilID, fmt.Errorf(f, a...)
	}

	data, err := doc.db().Create(string(NewID().Thing(doc.Table())), doc.doc)
	if err != nil {
		return errWrapf("sdb: create: %w", err)
	}

	var docId struct {
		RawId string `json:"id"`
	}

	if err := surrealdb.Unmarshal(data, &docId); err != nil {
		return errWrapf("unmarshal docId: %w", err)
	}

	recId, err := NewIDFromThing(Thing(docId.RawId), doc.Table())
	if err != nil {
		return errWrapf("new id from thing: %w", err)
	}

	return recId, nil
}
