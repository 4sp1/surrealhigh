package gold

import (
	"encoding/json"
	"fmt"
	surrealhigh "github.com/4sp1/surrealhigh"
	"time"
)

type A struct {
	S  string
	T  time.Time
	id surrealhigh.Id
	th surrealhigh.Thing
}
type docA struct {
	S     fDocA_S     `json:"s"`
	T     fDocA_T     `json:"t"`
	DocID fDocA_DocID `json:"id"`
}
type fDocA_S string
type fDocA_T fDocA_T_struct
type fDocA_T_struct struct {
	t time.Time
}

func (a A) doc() *docA {
	var doc docA
	doc.S = fDocA_S(a.S)
	doc.T = fDocA_T(fDocA_T_struct{t: a.T})
	return &doc
}

type fDocA_DocID surrealhigh.Id

func (doc docA) Id() surrealhigh.Thing {
	return surrealhigh.Id(doc.DocID).Thing(doc.Table())
}
func (doc docA) Table() surrealhigh.Table {
	return "a"
}
func (_ fDocA_S) Field() surrealhigh.Field {
	return "s"
}
func (_ fDocA_T) Field() surrealhigh.Field {
	return "t"
}
func (_ fDocA_DocID) Field() surrealhigh.Field {
	return "id"
}
func (_ fDocA_DocID) Table() surrealhigh.Table {
	return "a"
}
func (v *fDocA_T) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(v.t))
}
func (v *fDocA_T) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &v.t)
}
func (id fDocA_DocID) MarshalJSON() ([]byte, error) {
	sid := surrealhigh.Id(id)
	sth := sid.Thing(id.Table())
	return []byte("\"" + sth + "\""), nil
}
func (v fDocA_DocID) UnmarshalJSON(b []byte) error {
	th := surrealhigh.Thing(b[1 : len(b)-1])
	tb := docA{}.Table()
	id, err := surrealhigh.NewIDFromThing(th, tb)
	if err != nil {
		return fmt.Errorf("surrealhigh: new id from thing: %w", err)
	}
	for i := 0; i < 16; i++ {
		v[i] = id[i]
	}
	return nil
}
