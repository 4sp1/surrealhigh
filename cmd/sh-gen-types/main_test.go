package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	err := _main(
		[]string{"./test"}, []string{}, []string{"a"}, "model", "./test/model_gen.go")
	require.NoError(t, err)
}