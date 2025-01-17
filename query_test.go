package surrealhigh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQueryFrom(t *testing.T) {
	t.Run("select * from records", func(t *testing.T) {
		q := NewQueryFrom(Table("records"))
		assert.Equal(t, "SELECT * FROM records", q.String())
	})
	t.Run("select * from logs where record_id=$rid", func(t *testing.T) {
		q := NewQueryFrom(Table("logs"), QueryOptionWhere(
			NewConditionIs(NewConditionAtomField(Field("record_id")), NewConditionAtomVar("rid", nil))))
		assert.Equal(t, "SELECT * FROM logs WHERE (record_id IS $rid)", q.String())
	})
	t.Run("select * from logs order by timestamp asc", func(t *testing.T) {
		q := NewQueryFrom(Table("logs"), QueryOptionOrderByAsc(Field("timestamp")))
		assert.Equal(t, "SELECT * FROM logs ORDER BY timestamp ASC", q.String())
	})
	t.Run("select * from records where record_id=$id and is_out=$out", func(t *testing.T) {
		q := NewQueryFrom(Table("records"), QueryOptionWhere(
			NewConditionAnd(
				NewConditionIs(NewConditionAtomField(Field("record_id")), NewConditionAtomVar("id", nil)),
				NewConditionIs(NewConditionAtomField(Field("is_out")), NewConditionAtomVar("out", nil)),
			),
		))
		assert.Equal(t, "SELECT * FROM records WHERE ((record_id IS $id) AND (is_out IS $out))", q.String())
	})
}

func TestCondition_ValuedVars(t *testing.T) {
	for _, test := range []struct {
		name      string
		condition Condition
		vars      []conditionAtomVar
	}{
		{
			name: "record_id = $id(0) and is_out = $out(false)",
			condition: NewConditionAnd(
				NewConditionIs(NewConditionAtomField(Field("is_out")), NewConditionAtomVar("out", false)),
				NewConditionIs(NewConditionAtomField(Field("record_id")), NewConditionAtomVar("id", 0)),
			),
			vars: []conditionAtomVar{
				{name: varWhereClause("out"), value: false},
				{name: varWhereClause("id"), value: 0},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.vars, test.condition.valuedVars())
		})
	}
}

func TestWhereClauseString(t *testing.T) {
	t.Run("a0=$a0 and (b1=$b1 or c2=$c2) and d3=$d3 and e4=$e4", func(t *testing.T) {
		assert.Equal(t, "((a0 IS $a0) AND (((b1 IS $b1) OR (c2 IS $c2)) AND ((d3 IS $d3) AND (e4 IS $e4))))", NewConditionAnd(
			NewConditionIs(NewConditionAtomField(Field("a0")), NewConditionAtomVar("a0", nil)),
			NewConditionOr(
				NewConditionIs(NewConditionAtomField(Field("b1")), NewConditionAtomVar("b1", nil)),
				NewConditionIs(NewConditionAtomField(Field("c2")), NewConditionAtomVar("c2", nil)),
			),
			NewConditionIs(NewConditionAtomField(Field("d3")), NewConditionAtomVar("d3", nil)),
			NewConditionIs(NewConditionAtomField(Field("e4")), NewConditionAtomVar("e4", nil)),
		).String())
	})
	t.Run("key=$key and record_id=$id and is_out=false", func(t *testing.T) {
		keyField, recordIdField, isOutField := Field("key"), Field("record_id"), Field("is_out")
		keyVar, idVar := varWhereClause("key"), varWhereClause("id")
		keyIsKeyClause := binaryWhereClause{
			l:  fieldWhereClause(keyField),
			op: whereOpIs,
			r:  keyVar,
		}
		recordIdIsIdClause := binaryWhereClause{
			l:  fieldWhereClause(recordIdField),
			op: whereOpIs,
			r:  idVar,
		}
		isOutFalseClause := binaryWhereClause{
			l:  fieldWhereClause(isOutField),
			op: whereOpIs,
			r:  boolWhereClause(false),
		}
		clause := binaryWhereClause{
			l:  keyIsKeyClause,
			op: whereOpAnd,
			r: binaryWhereClause{
				l:  recordIdIsIdClause,
				op: whereOpAnd,
				r:  isOutFalseClause,
			},
		}
		assert.Equal(t, "((key IS $key) AND ((record_id IS $id) AND (is_out IS false)))", clause.String())
	})
}

func TestQueryString(t *testing.T) {
	t.Run("select * from records where (record_id = $id) order by timestamp asc", func(t *testing.T) {
		var (
			recordIdField, timestampField = Field("record_id"), Field("timestamp")
			idVar                         = varWhereClause("id")
			recordsTable                  = Table("records")
		)
		q := selectStatement{
			where: binaryWhereClause{
				l:  recordIdField,
				op: whereOpIs,
				r:  idVar,
			},
			orderBy: &selectOrderBy{
				order: selectOrderAsc,
				field: timestampField,
			},
			from: recordsTable,
		}
		assert.Equal(t, "SELECT * FROM records WHERE (record_id IS $id) ORDER BY timestamp ASC", q.String())
	})
}
