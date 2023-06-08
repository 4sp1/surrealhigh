package surrealhigh

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

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

var (
	ErrNotInThisTable = errors.New("thing not in this table")
	ErrBadThing       = errors.New("bad thing")
)

func NewIDFromThing(th Thing, tb Table) (_ Id, err error) {
	rawId, found := strings.CutPrefix(tb.Prefix(), string(th))
	if !found {
		fields := strings.FieldsFunc(th.String(), func(r rune) bool { return r == ':' })
		if len(fields) != 2 {
			err = fmt.Errorf("strings: len(colon fields)!=2: %w", ErrBadThing)
			return
		}
		err = fmt.Errorf("strings: cut prefix: %w: %q not in %q; in %q", ErrNotInThisTable, string(th), string(tb), fields[0])
		return
	}
	rawId = strings.ReplaceAll(rawId, "_", "-")
	uid, err := uuid.Parse(rawId)
	if err != nil {
		log.Debug().Err(err).Msg("uuid: parse: " + rawId)
		err = fmt.Errorf("uuid: parse: %w", err)
		return
	}
	return Id(uid), nil
}

func (i Id) String() string {
	canonic := strings.ReplaceAll(uuid.UUID(i).String(), "-", "_")
	return canonic
}

type Table string

func (t Table) Prefix() string {
	return string(t) + ":"
}

func (i Id) Thing(t Table) Thing {
	return Thing(t.Prefix() + i.String())
}

func NewID() Id {
	return Id(uuid.New())
}
