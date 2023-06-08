package surrealhigh

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestId_Thing(t *testing.T) {
	t.Run("null id", func(t *testing.T) {
		id := Id(uuid.Nil)
		table := Table("test")
		assert.Equal(t, Thing("test:00000000_0000_0000_0000_000000000000"), id.Thing(table))
	})
}

func TestThing_String(t *testing.T) {
	assert.Equal(t, "test", Thing("test").String())
}

func TestNewID(t *testing.T) {
	id := NewID()
	assert.Equal(t, strings.ReplaceAll(uuid.UUID(id).String(), "-", "_"), id.String())
}

func TestNewIDFromThing(t *testing.T) {
	for _, test := range []struct {
		v           Id
		err         error
		errContains string
		name        string
		th          Thing
		tb          Table
	}{
		{
			v:    Id(uuid.Nil),
			name: "match",
			th:   "test:00000000_0000_0000_0000_000000000000",
			tb:   "test",
		},
		{
			err:  ErrNotInThisTable,
			name: "not this table",
			th:   "test:00000000_0000_0000_0000_000000000000",
			tb:   "taste",
		},
		{
			err:  ErrBadThing,
			name: "thing colon missing",
			th:   "test00000000_0000_0000_0000_000000000000",
			tb:   "test",
		},
		{
			errContains: "",
			name:        "thing colon missing",
			th:          "test00000000_0000_0000_0000_000000000000",
			tb:          "test",
		},
	} {
		assert := assert.New(t)
		t.Run(test.name, func(t *testing.T) {
			id, err := NewIDFromThing(test.th, test.tb)
			if test.err != nil {
				assert.ErrorIs(err, test.err)
				return
			}
			if test.errContains != "" {
				assert.ErrorContains(err, test.errContains)
				return
			}
			if err != nil && test.err != nil {
				assert.NoError(err)
				return
			}
			assert.Equal(test.v, id)
		})
	}
}
