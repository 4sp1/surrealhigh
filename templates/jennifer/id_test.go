package jennifer

import (
	"fmt"
	"testing"

	"github.com/4sp1/surrealhigh"
	"github.com/stretchr/testify/assert"
)

func TestDoc_Ids(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		table surrealhigh.Table
		field surrealhigh.Field
		t     string
		id    string
		fn    func(surrealhigh.Table, surrealhigh.Field, string) string
	}{
		// fDoc${Table}_${Field}
		{
			table: "test",
			field: "fieldName",
			id:    "fDocTest_FieldName",
			fn: func(table surrealhigh.Table, field surrealhigh.Field, t string) string {
				return DocField{Field: field}.
					docStructFieldTypeId(Doc{table: table})
			},
		},
		// doc${Table}
		{
			table: "test",
			id:    "docTest",
			fn: func(table surrealhigh.Table, field surrealhigh.Field, t string) string {
				return Doc{table: table}.docStructId()
			},
		},
		// `json:"${field}"`
		{
			field: "test",
			id:    "test",
			fn: func(table surrealhigh.Table, field surrealhigh.Field, t string) string {
				return DocField{Field: field, t: t}.Tag()["json"]
			},
		},
		// ${Field}
		{
			field: "test",
			id:    "Test",
			fn: func(table surrealhigh.Table, field surrealhigh.Field, t string) string {
				return DocField{Field: field, t: t}.docStructFieldNameId()
			},
		},
		// fDoc${Table}_DocID
		{
			table: "test",
			id:    "fDocTest_DocID",
			fn: func(table surrealhigh.Table, field surrealhigh.Field, t string) string {
				return Doc{table: table}.docIdType()
			},
		},
		// ${Table}
		{
			table: "test",
			id:    "Test",
			fn: func(table surrealhigh.Table, field surrealhigh.Field, t string) string {
				return Doc{table: table}.docPublicId()
			},
		},
	} {
		t.Run(test.id, func(t *testing.T) {
			assert.Equal(t, test.id, test.fn(test.table, test.field, test.t))
		})
	}
}

func TestDoc_docStructFields(t *testing.T) {
	t.Run("len+1", func(t *testing.T) {
		fields := make([]DocField, 2)
		for i := 0; i < 2; i++ {
			fields[i] = NewField(fmt.Sprintf("a_%d", i), "test")
		}
		assert.Equal(t, 3, len(Doc{fields: fields}.docStructFields()))
	})
}