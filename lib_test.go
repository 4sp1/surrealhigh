package surrealhigh

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestId_Thing(t *testing.T) {
	t.Run("null id", func(t *testing.T) {
		id := Id(uuid.Nil)
		table := Table("test")
		assert.Equal(t, Thing("test:⟨00000000-0000-0000-0000-000000000000⟩"), id.Thing(table))
	})
}

func TestThing_String(t *testing.T) {
	assert.Equal(t, "test", Thing("test").String())
}

func TestNewID(t *testing.T) {
	id := NewID()
	assert.Equal(t, uuid.UUID(id).String(), id.String())
}
