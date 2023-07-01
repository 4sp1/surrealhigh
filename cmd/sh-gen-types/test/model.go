package model

import "time"

//go:generate sh-gen-types -doc=a -pkg model -o model_gen.go
type a struct {
	t time.Time
	b []byte
	s string
	p *time.Time

	// TODO(malikbenkirane) support for special time cases
	// lp []*time.Time
	// ls []time.Time
}