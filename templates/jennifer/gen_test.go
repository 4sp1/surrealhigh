package jennifer

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

const dir = "gold"

func TestNewDoc(t *testing.T) {
	t.Run("doc A", func(t *testing.T) {
		_, filename, _, _ := runtime.Caller(0)
		cd := path.Dir(filename)
		dirPath := path.Join(cd, dir)
		if _, err := os.Stat(dirPath); err != nil {
			err := os.Mkdir(dirPath, 0755)
			require.NoError(t, err, "prepare dirPath")
		}
		output := NewDoc("gold", "a", NewField("s", "string")).file
		f, err := os.Create(path.Join(dirPath, "sigh_doc_gen.go"))
		require.NoError(t, err, "create sigh_doc_gen.go")
		require.NoError(t, output.Render(f))
		require.NoError(t, err, f.Close())
		// todo)) compile and require no error
	})
}
