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
		if field.qual != "" {
			pubDocFields = append(pubDocFields, Id(field.docStructFieldNameId()).Qual(field.qual, field.t))
			continue
		}
		pubDocFields = append(pubDocFields, Id(field.docStructFieldNameId()).Id(field.t))
	}
	pubDocFields = append(pubDocFields,
		Id("id").Qual(origin, "Id"),
		Id("th").Qual(origin, "Thing"))
	f.Type().Id(cc(table.String())).Struct(pubDocFields...)

	// ## public doc to private doc
	// func (a A) doc() *docA {...}

	{
		a := string(table)
		block := []Code{Var().Id("doc").Id(doc.docStructId())}
		for _, field := range fields {
			block = append(block,
				Id("doc").Dot(field.docStructFieldNameId()).
					Op("=").
					Id(field.docStructFieldTypeId(doc)).Parens(
					Id(a).Dot(field.docStructFieldNameId())))
		}
		block = append(block, Return(Op("&").Id("doc")))
		f.Func().Params(Id(a).Id(doc.docPublicId())).
			Id("doc").Params().Op("*").Id(doc.docStructId()).
			Block(block...)
	}

	// ## doc struct type
	// type docA struct {...}

	f.Type().Id(doc.docStructId()).
		Struct(doc.docStructFields()...)

	// ## field types
	// type fDocA_S t

	for _, field := range fields {
		stmt := f.Type().Id(field.docStructFieldTypeId(doc))
		if field.isptr {
			stmt = stmt.Op("*")
		}
		if field.qual == "" {
			stmt.Id(field.t)
			continue
		}
		stmt.Qual(field.qual, field.t)
	}

	// ## doc id type surrealhihg.Id
	// type fDocA_DocId surrealhigh.Id

	f.Type().Id(doc.docIdType()).Qual(origin, "Id")

	// ## doc.Id() method
	// func (doc docA) Id() surrealhigh.Id { return doc.DocID }

	f.Func().
		Params(Id("doc").Id(doc.docStructId())).
		Id("Id").
		Params().
		Qual(origin, "Id").
		Block(
			Return(Qual(origin, "Id").Parens(Id("doc").Dot("DocID"))))

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

	// DocID marshaler
	// func (id fDocA_DocID) MarshalJSON() ([]byte, error) {...}

	f.Func().Params(Id("id").Id(doc.docIdType())).
		Id("MarshalJSON").
		Params().
		Params(Index().Byte(), Error()).
		Block(
			Id("sid").Op(assign).Qual(origin, "Id").Parens(Id("id")),
			Id("sth").Op(assign).Id("sid").Dot("Thing").Call(Id("id").Dot("Table").Call()),
			Return(Index().Byte().Parens(
				Lit(litQuote).Op(plus).Id("sth").Op(plus).Lit(litQuote),
			), Nil()))

	// DocID unmarshaler
	// func (id fDocA_DocID) UnmarshalJSON(b []byte) error {...}

	f.Func().Params(Id("v").Id(doc.docIdType())).
		Id("UnmarshalJSON").
		Params(Id("b").Index().Byte()).
		Params(Error()).
		Block(
			Id("tb").Op(assign).Qual(origin, "Table").Parens(Lit("z")),
			Id("th").Op(assign).Qual(origin, "Thing").Parens(
				Lit("z:").Op(plus).String().Parens(Id("b"))),
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
