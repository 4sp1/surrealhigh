package surrealhigh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO move to a sub test package

type mockDriver struct{}

func (driver mockDriver) driver() surrealDriver {
	return mockDriverResult{}
}

type mockDriverResult struct{}

func (mock mockDriverResult) Query(sql string, vars interface{}) (interface{}, error) {

	// 	[
	// >   {
	// >     "result": [ {...}, ... ],
	// >     "status": "OK",
	// >     "time": "337.295µs"
	// >   }
	// > ]

	var results []interface{}
	results = append(results, struct {
		Result []mockDoc `json:"result"`
		Status string    `json:"status"`
	}{
		Result: []mockDoc{
			{RecordID: 0, IsOut: false},
		},
		Status: "OK",
	})

	return results, nil

}

type mockDoc struct {
	RecordID int
	IsOut    bool
}

func (doc mockDoc) Table() Table {
	return Table("")
}

func TestDBSelect_Do(t *testing.T) {
	t.Run("mock driver result no error", func(t *testing.T) {
		results, err := SelectOn[mockDoc](NewQueryFrom(Table("")), mockDriver{}).Do()
		require.NoError(t, err)
		assert.Equal(t, []mockDoc{
			{RecordID: 0, IsOut: false},
		}, results)
	})

}
