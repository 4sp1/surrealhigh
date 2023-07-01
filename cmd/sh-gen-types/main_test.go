package main

import (
	"testing"

	"github.com/4sp1/surrealhigh/templates/jennifer"
	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	err := jennifer.NewGen(
		[]string{"./test"}, []string{}, []string{"a"}, "model", "./test/model_gen.go")
	require.NoError(t, err)
}