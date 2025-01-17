package surrealhigh

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

func NewQueryFrom(from Table, opts ...QueryOption) Select {
	q := Select{
		valuedSelectStatement: valuedSelectStatement{
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

type Select struct{ valuedSelectStatement }

var _ valuedWhereClause = valuedSelectStatement{}

func (vc valuedSelectStatement) String() string {
	return vc.asWhereClause().String()
}

func (vc valuedSelectStatement) asWhereClause() whereClause {
	c := selectStatement{
		orderBy: vc.orderBy,
		from:    vc.from,
	}
	if vc.where != nil {
		c.where = vc.where.asWhereClause()
	}
	return c
}

func (vc valuedSelectStatement) valuedVars() []conditionAtomVar {
	return vc.where.valuedVars()
}

type valuedSelectStatement struct {
	orderBy *selectOrderBy
	where   valuedWhereClause
	from    Table
}

type selectStatement struct {
	orderBy *selectOrderBy
	from    Table
	where   whereClause
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
	whereClause

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

func NewConditionAnd(c0 Condition, c ...Condition) Condition {
	return newRecursiveCondition(whereOpAnd, c0, c...)
}

func NewConditionOr(c0 Condition, c ...Condition) Condition {
	return newRecursiveCondition(whereOpOr, c0, c...)
}

func newBinaryCondition(l Condition, r Condition, op whereOp) Condition {
	return valuedBinaryWhereClause{op: op, l: l, r: r}
}

func newRecursiveCondition(op whereOp, c0 Condition, c ...Condition) Condition {
	if len(c) == 0 {
		return c0
	}
	if len(c) == 1 {
		return newBinaryCondition(c0, c[0], op)
	}
	return newBinaryCondition(c0, newRecursiveCondition(op, c[0], c[1:]...), op)
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
	whereOpOr    = whereOp("OR")
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
