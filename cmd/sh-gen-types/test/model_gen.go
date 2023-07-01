package model

import (
	"encoding/json"
	"fmt"
	surrealhigh "github.com/4sp1/surrealhigh"
	"time"
)

type A struct {
	T  time.Time
	B  []byte
	S  string
	P  *time.Time
	id surrealhigh.Id
	th surrealhigh.Thing
}
type docA struct {
	T     fDocA_T     `json:"t"`
	B     fDocA_B     `json:"b"`
	S     fDocA_S     `json:"s"`
	P     fDocA_P     `json:"p"`
	DocID fDocA_DocID `json:"id"`
}
type fDocA_T fDocA_T_struct
type fDocA_B []byte
type fDocA_S string
type fDocA_P fDocA_P_struct
type fDocA_T_struct struct {
	t time.Time
}
type fDocA_P_struct struct {
	t *time.Time
}

func (a A) doc() *docA {
	var doc docA
	doc.T = fDocA_T(fDocA_T_struct{t: a.T})
	doc.B = fDocA_B(a.B)
	doc.S = fDocA_S(a.S)
	doc.P = fDocA_P(fDocA_P_struct{t: a.P})
	return &doc
}

type fDocA_DocID surrealhigh.Id

func (doc docA) Id() surrealhigh.Thing {
	return surrealhigh.Id(doc.DocID).Thing(doc.Table())
}
func (doc docA) Table() surrealhigh.Table {
	return "a"
}
func (_ fDocA_T) Field() surrealhigh.Field {
	return "t"
}
func (_ fDocA_B) Field() surrealhigh.Field {
	return "b"
}
func (_ fDocA_S) Field() surrealhigh.Field {
	return "s"
}
func (_ fDocA_P) Field() surrealhigh.Field {
	return "p"
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
func (v *fDocA_P) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*v.t))
}
func (v *fDocA_P) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, v.t)
}
func (id *fDocA_DocID) MarshalJSON() ([]byte, error) {
	sid := surrealhigh.Id(*id)
	sth := sid.Thing(id.Table())
	return []byte("\"" + sth + "\""), nil
}
func (v *fDocA_DocID) UnmarshalJSON(b []byte) error {
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
