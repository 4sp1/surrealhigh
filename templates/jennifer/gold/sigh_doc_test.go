package gold

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/4sp1/surrealhigh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeEmbedding(t *testing.T) {
	a := A{T: time.Now()}
	adoc := a.doc()
	adoc.DocID = fDocA_DocID(surrealhigh.NewID())
	b, err := json.Marshal(adoc)
	require.NoError(t, err)
	fmt.Println(string(b))
	var u docA
	err = json.Unmarshal(b, &u)
	require.NoError(t, err)
	assert.Equal(t, a.T.Format(time.RFC3339), u.T.t.Format(time.RFC3339))
}