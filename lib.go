package surrealhigh

import "github.com/google/uuid"

type Field string

func (f Field) String() string {
	return string(f)
}

// Thing is a pointer to a Doc
type Thing string

type Id uuid.UUID

func (t Thing) String() string {
	return string(t)
}

func (i Id) String() string {
	return uuid.UUID(i).String()
}

type Table string

func (t Table) Prefix() string {
	return string(t) + ":"
}

func (i Id) Thing(t Table) Thing {
	return Thing(t.Prefix() + "`" + i.String() + "`")
}

func NewID() Id {
	return Id(uuid.New())
}
