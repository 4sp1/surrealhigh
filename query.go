package surrealhigh

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

func NewQueryFrom(from Table, opts ...QueryOption) Select {
	q := Select{
		selectStatement: selectStatement{
			from: from,
		},
	}
	for _, opt := range opts {
		q = opt(q)
	}
	return q
}

type QueryOption func(Select) Select

func QueryOptionWhere(c Condition) QueryOption {
	return func(q Select) Select {
		q.where = c
		return q
	}
}

func QueryOptionOrderByAsc(f Field) QueryOption {
	return func(q Select) Select {
		q.orderBy = &selectOrderBy{
			field: f,
			order: selectOrderAsc,
		}
		return q
	}
}

func QueryOptionOrderByDesc(f Field) QueryOption {
	return func(q Select) Select {
		q.orderBy = &selectOrderBy{
			field: f,
			order: selectOrderDesc,
		}
		return q
	}
}

type Select struct{ selectStatement }

type selectStatement struct {
	orderBy *selectOrderBy
	where   whereClause
	from    Table
}

type (
	Condition interface {
		whereClause
		valuedWhereClause
	}
	ConditionAtom interface{ valuedWhereClause }

	ConditionAtomField interface{ ConditionAtom }
	ConditionAtomVar   interface{ ConditionAtom }
)

var (
	_ ConditionAtomField = fieldWhereClause("")
	_ ConditionAtomVar   = conditionAtomVar{}
	_ Condition          = valuedBinaryWhereClause{}
)

type valuedWhereClause interface {
	valuedVars() []conditionAtomVar
	asWhereClause() whereClause
}

type valuedBinaryWhereClause struct {
	l  valuedWhereClause
	r  valuedWhereClause
	op whereOp
}

var (
	_ valuedWhereClause = fieldWhereClause("")
	_ valuedWhereClause = valuedBinaryWhereClause{}
	_ valuedWhereClause = conditionAtomVar{}
)

func (vc valuedBinaryWhereClause) asWhereClause() whereClause {
	return binaryWhereClause{
		l:  vc.l.asWhereClause(),
		r:  vc.r.asWhereClause(),
		op: vc.op,
	}
}

func (vc valuedBinaryWhereClause) String() string {
	return vc.asWhereClause().String()
}

func (vc fieldWhereClause) asWhereClause() whereClause {
	return vc
}

func (vc conditionAtomVar) asWhereClause() whereClause {
	return vc.name
}

func (c valuedBinaryWhereClause) valuedVars() (vars []conditionAtomVar) {
	vars = append(vars, c.l.valuedVars()...)
	vars = append(vars, c.r.valuedVars()...)
	return vars
}

func (c fieldWhereClause) valuedVars() []conditionAtomVar {
	return []conditionAtomVar{}
}

func (c conditionAtomVar) valuedVars() []conditionAtomVar {
	return []conditionAtomVar{c}
}

func NewConditionAtomField(f Field) ConditionAtomField {
	return fieldWhereClause(f)
}

func NewConditionAtomVar(name string, value interface{}) ConditionAtomVar {
	return conditionAtomVar{
		name:  varWhereClause(name),
		value: value,
	}
}

type conditionAtomVar struct {
	name  varWhereClause
	value interface{}
}

func (v conditionAtomVar) String() string {
	return v.name.String()
}

func NewConditionIs(l ConditionAtom, r ConditionAtom) Condition {
	return valuedBinaryWhereClause{l: l, r: r, op: whereOpIs}
}

func NewConditionIsNot(l ConditionAtom, r ConditionAtom) Condition {
	return valuedBinaryWhereClause{l: l, r: r, op: whereOpIsNot}
}

func NewConditionAnd(l Condition, r Condition) Condition {
	return valuedBinaryWhereClause{
		op: whereOpAnd,
		l:  l,
		r:  r,
	}
}

type selectOrderBy struct {
	order selectOrder
	field Field
}

type selectOrder string

const (
	selectOrderDesc = selectOrder("DESC")
	selectOrderAsc  = selectOrder("ASC")
)

type whereClause interface {
	fmt.Stringer
}

func (q selectStatement) String() string {
	b := strings.Builder{}
	b.WriteString("SELECT * FROM ")
	b.WriteString(string(q.from))
	if q.where != nil {
		b.WriteString(" WHERE ")
		b.WriteString(q.where.String())
	}
	if q.orderBy != nil {
		b.WriteString(" ORDER BY ")
		b.WriteString(q.orderBy.field.String())
		b.WriteString(" ")
		b.WriteString(string(q.orderBy.order))
	}
	return b.String()
}

func (w binaryWhereClause) String() string {
	return fmt.Sprintf("(%s %s %s)", w.l.String(), w.op, w.r.String())
}

type whereOp string

const (
	whereOpAnd   = whereOp("AND")
	whereOpIs    = whereOp("IS")
	whereOpIsNot = whereOp("IS NOT")
	// TODO https://surrealdb.com/docs/surrealql/operators
)

type binaryWhereClause struct {
	l  whereClause
	r  whereClause
	op whereOp

	log *zerolog.Logger
}

var (
	_ whereClause = binaryWhereClause{}
	_ whereClause = fieldWhereClause(Field(""))
	_ whereClause = varWhereClause("")
	_ whereClause = boolWhereClause(false)
)

type boolWhereClause bool

func (c boolWhereClause) String() string {
	if c {
		return "true"
	}
	return "false"
}

type fieldWhereClause Field

func (c fieldWhereClause) String() string {
	return Field(c).String()
}

type varWhereClause string

func (c varWhereClause) Var() string {
	return string(c)
}

func (c varWhereClause) String() string {
	return "$" + string(c)
}
