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
			ConditionIs(NewConditionAtomField(Field("record_id")), NewConditionAtomVar("rid", nil))))
		assert.Equal(t, "SELECT * FROM logs WHERE (record_id IS $rid)", q.String())
	})
	// TODO t.Run("select * from records where state=$state and date<$until")
	t.Run("select * from logs order by timestamp asc", func(t *testing.T) {
		q := NewQueryFrom(Table("logs"), QueryOptionOrderByAsc(Field("timestamp")))
		assert.Equal(t, "SELECT * FROM logs ORDER BY timestamp ASC", q.String())
	})
}

func TestWhereClauseString(t *testing.T) {
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
		q := query{
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
