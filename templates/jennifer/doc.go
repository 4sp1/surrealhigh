package jennifer

import (
	"io"
	"strings"
	"unicode"

	sh "github.com/4sp1/surrealhigh"
	. "github.com/dave/jennifer/jen"
	"github.com/rs/zerolog/log"
)

const origin = "github.com/4sp1/surrealhigh"

type Doc struct {
	table  sh.Table
	fields []DocField

	file *File
}

func (doc Doc) Write(w io.Writer) error {
	return doc.file.Render(w)
}

type DocField struct {
	sh.Field

	t     string
	qual  string
	isptr bool
	isarr bool
}

func (f DocField) isTime() bool {
	return f.t == "Time" && f.qual == "time"
}

type NewFieldOption func(DocField) DocField

func NewFieldWithQual(qual string) NewFieldOption {
	log.Trace().Str("qual", qual).Msg("NewFieldWithQual")
	return func(df DocField) DocField {
		df.qual = qual
		return df
	}
}

func NewFieldWithPointer() NewFieldOption {
	log.Trace().Msg("NewFieldWithPointer")
	return func(df DocField) DocField {
		df.isptr = true
		return df
	}
}

func NewFieldWithArray() NewFieldOption {
	return func(df DocField) DocField {
		df.isarr = true
		return df
	}
}

func NewField(name, t string, opts ...NewFieldOption) DocField {
	f := DocField{Field: sh.Field(name), t: t}
	log.Trace().Str("name", name).Str("t", t).Msg("NewField")
	for _, opt := range opts {
		f = opt(f)
	}
	return f
}

// doc${Table}
func (doc Doc) docStructId() string {
	return "doc" + cc(doc.table.String())
}

// f${Table}_${Field}
func (field DocField) docStructFieldTypeId(doc Doc) string {
	return "f" + cc(doc.docStructId()) + "_" + cc(field.String())
}

// f${Table}_${Field}_struct
func (field DocField) docStructFieldTypeStructTypeId(doc Doc) string {
	return "f" + cc(doc.docStructId()) + "_" + cc(field.String()) + "_struct"
}

// ${Field}
func (field DocField) docStructFieldNameId() string {
	return cc(field.String())
}

// `json:"${field}"`
func (field DocField) Tag() map[string]string {
	return map[string]string{"json": field.String()}
}

func (doc Doc) docStructFields() (codes []Code) {
	for _, field := range doc.fields {
		// XXX
		// if field.qual != "" {
		// 	codes = append(codes,
		// 		Id(field.docStructFieldNameId()).
		// 			Qual(field.qual, field.docStructFieldTypeId(doc)).
		// 			Tag(field.Tag()))
		// 	continue
		// }
		codes = append(codes,
			Id(field.docStructFieldNameId()).
				Id(field.docStructFieldTypeId(doc)).
				Tag(field.Tag()))
	}
	return append(codes,
		Id("DocID").Id(doc.docIdType()).Tag(map[string]string{
			"json": "id",
		}))
}

func (doc Doc) docPublicId() string { // ${Table}
	return cc(doc.table.String())
}

const docIdField = "id"

func (doc Doc) docIdType() string { // f${Table}_DocID
	return "fDoc" + cc(doc.table.String()) + "_DocID"
}

func NewDoc(pkg sh.Package, table sh.Table, fields ...DocField) (doc Doc) {

	doc.table = table
	doc.fields = fields // DocId field not included and treated separate

	// ## package
	// package ...

	f := NewFile(pkg.String())

	// ## public doc struct type
	// type A struct {...}

	var pubDocFields []Code
	for _, field := range fields {
		stmt := Id(field.docStructFieldNameId())
		if field.isarr {
			stmt = stmt.Index()
		}
		if field.isptr {
			stmt = stmt.Op("*")
		}
		if field.qual != "" {
			stmt = stmt.Qual(field.qual, field.t)
		} else {
			stmt = stmt.Id(field.t)
		}
		pubDocFields = append(pubDocFields, stmt)
	}
	pubDocFields = append(pubDocFields,
		Id("id").Qual(origin, "Id"),
		Id("th").Qual(origin, "Thing"))
	f.Type().Id(cc(table.String())).Struct(pubDocFields...)

	// ## doc struct type
	// type docA struct {...}

	f.Type().Id(doc.docStructId()).
		Struct(doc.docStructFields()...)

	// ## field types
	// type fDocA_S t
	times := make(map[DocField]struct{})
	for _, field := range fields {
		stmt := f.Type().Id(field.docStructFieldTypeId(doc))
		if !field.isTime() {
			if field.isarr {
				stmt = stmt.Index()
			}
			if field.isptr {
				stmt = stmt.Op("*")
			}
		}
		if field.qual == "" {
			stmt.Id(field.t)
			continue
		}
		// TODO(malikbenkirane) this could work for any struct type
		// suggested naming "embed", "structure embedding", etc...
		if field.isTime() {
			t := field.docStructFieldTypeStructTypeId(doc)
			stmt.Id(t)
			times[field] = struct{}{}
			continue
		}
		stmt.Qual(field.qual, field.t)
	}
	for t := range times {
		stmt := Id("t")
		if t.isarr {
			stmt = stmt.Index()
		}
		if t.isptr {
			stmt = stmt.Op("*")
		}
		stmt = stmt.Qual("time", "Time")
		f.Type().
			Id(t.docStructFieldTypeStructTypeId(doc)).
			Struct(stmt)
	}

	// ## public doc to private doc
	// func (a A) doc() *docA {...}

	{
		a := string(table)
		block := []Code{Var().Id("doc").Id(doc.docStructId())}
		for _, field := range fields {
			stmt := Id("doc").Dot(field.docStructFieldNameId()).
				Op("=").Id(field.docStructFieldTypeId(doc))
			if _, ok := times[field]; ok {
				stmt = stmt.Parens(
					Id(field.docStructFieldTypeStructTypeId(doc)).
						Values(Dict{Id("t"): Id(a).Dot(field.docStructFieldNameId())}))
			} else {
				stmt = stmt.Parens(Id(a).Dot(field.docStructFieldNameId()))
			}
			block = append(block, stmt)
		}
		block = append(block, Return(Op("&").Id("doc")))
		f.Func().
			Params(Id(a).Id(doc.docPublicId())). // (a A)
			Id("doc").                           // doc
			Params().                            // ()
			Op("*").Id(doc.docStructId()).       // *docA
			Block(block...)                      // {...}
	}

	// ## doc id type surrealhihg.Id
	// type fDocA_DocId surrealhigh.Id

	f.Type().Id(doc.docIdType()).Qual(origin, "Id")

	// ## doc.Id() method
	// func (doc docA) Id() surrealhigh.Thing { return surrealhigh.Id(doc.DocID).Thing(doc.Table()) }

	f.Func().
		Params(Id("doc").Id(doc.docStructId())).
		Id("Id").
		Params().
		Qual(origin, "Thing").
		Block(
			Return(Qual(origin, "Id").Parens(Id("doc").Dot("DocID"))).Dot("Thing").Call(Id("doc").Dot("Table").Call()))

	// ## doc.Table() method
	// func (doc docA) Table() surrealhigh.Table { return "a" }

	{
		litTable := Lit(doc.table.String())
		f.Func().
			Params(Id("doc").Id(doc.docStructId())).
			Id("Table").
			Params().
			Qual(origin, "Table").
			Block(
				Return(litTable))
	}

	// ## fields Field() method
	// func (_ fDocA_S) Field() surrealhigh.Field { return "s" }

	for _, field := range fields {
		litField := Lit(field.Field.String())
		f.Func().
			Params(Id("_").Id(field.docStructFieldTypeId(doc))).
			Id("Field").
			Params().
			Qual(origin, "Field").
			Block(
				Return(litField))
	}

	// ## DocID Field() method
	// func (_ fDocA_DocId) Field() surrealhigh.Field { return "id" }

	f.Func().
		Params(Id("_").Id(doc.docIdType())).
		Id("Field").
		Params().
		Qual(origin, "Field").
		Block(
			Return(Lit(docIdField)))

	// ## DocID Table() method
	// func (_ fDocA_DocID) Table() surrealhigh.Table { return "a" }

	{
		litTable := Lit(doc.table.String())
		f.Func().
			Params(Id("_").Id(doc.docIdType())).
			Id("Table").
			Params().
			Qual(origin, "Table").
			Block(
				Return(litTable))
	}

	// Times marshaler/unmarshaler TODO(malikbenkirane) read comments above
	//	func (v f${Table}_${Field}) MarshalJSON() ([]byte, error) {
	//  	return json.Marshal(time.Time(*v.t))
	// }
	//	func (v f${Table}_${Field}) UnmarshalJSON(b []byte) error {
	//  	return json.Umarshal(b, v.t)
	// }
	for t := range times {
		stmt := Id("v").Dot("t")
		if t.isptr {
			stmt = Op("*").Id("v").Dot("t")
		}
		f.Func().Params(Id("v").Op("*").Id(t.docStructFieldTypeId(doc))).
			Id("MarshalJSON").
			Params().
			Params(Index().Byte(), Error()).
			Block(Return(Qual("encoding/json", "Marshal").Call(Qual("time", "Time").Parens(stmt))))
		stmt = Op("&").Id("v").Dot("t")
		if t.isptr {
			stmt = Id("v").Dot("t")
		}
		f.Func().Params(Id("v").Op("*").Id(t.docStructFieldTypeId(doc))).
			Id("UnmarshalJSON").
			Params(Id("b").Index().Byte()).
			Params(Error()).
			Block(Return(Qual("encoding/json", "Unmarshal").Call(Id("b"), stmt)))
	}

	// DocID marshaler
	// func (id fDocA_DocID) MarshalJSON() ([]byte, error) {...}

	f.Func().
		Params(Id("id").Op("*").Id(doc.docIdType())).
		Id("MarshalJSON").
		Params().
		Params(Index().Byte(), Error()).
		Block(
			Id("sid").Op(assign).Qual(origin, "Id").Parens(Op("*").Id("id")),
			Id("sth").Op(assign).Id("sid").Dot("Thing").Call(Id("id").Dot("Table").Call()),
			Return(Index().Byte().Parens(
				Lit(litQuote).Op(plus).Id("sth").Op(plus).Lit(litQuote),
			), Nil()))

	// DocID unmarshaler
	// func (id fDocA_DocID) UnmarshalJSON(b []byte) error {...}

	f.Func().
		Params(Id("v").Op("*").Id(doc.docIdType())).
		Id("UnmarshalJSON").
		Params(Id("b").Index().Byte()).
		Params(Error()).
		Block(
			Id("th").Op(assign).
				Qual(origin, "Thing").Parens(
				Id("b").Index(Lit(1), Id("len").Call(Id("b")).Op("-").Lit(1))),
			Id("tb").Op(assign).
				Id(doc.docStructId()).Values().Dot("Table").Call(),
			List(Id("id"), Id("err")).Op(assign).
				Qual(origin, "NewIDFromThing").Call(Id("th"), Id("tb")),
			If(Id("err").Op(notEqual).Nil()).Block(
				Return(Qual("fmt", "Errorf").Call(
					Lit("surrealhigh: new id from thing: %w"), Id("err")))),
			For(
				Id("i").Op(assign).Lit(0),
				Id("i").Op(lt).Lit(16),
				Id("i").Op(pp),
			).Block(Id("v").Index(Id("i")).Op("=").Id("id").Index(Id("i"))),
			Return(Nil()))

	doc.file = f

	return doc
}

func cc(s string) string {
	var first bool
	return strings.Map(func(r rune) rune {
		if !first {
			first = true
			return unicode.ToUpper(r)
		}
		return r
	}, s)
}

const (
	assign   = ":="
	notEqual = "!="
	plus     = "+"
	lt       = "<"
	pp       = "++"
	ptr      = "*"

	litQuote = `"`
)
