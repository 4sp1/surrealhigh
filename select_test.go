package surrealhigh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO move to a sub test package

type mockDriver struct{ update *bool }

func newMockDriver() mockDriver {
	var b bool
	return mockDriver{&b}
}

func (driver mockDriver) driver() surrealDriver {
	return mockDriverResult{update: driver.update}
}

type mockDriverResult struct{ update *bool }

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
			{RecordID: 0, IsOut: *mock.update},
		},
		Status: "OK",
	})

	return results, nil

}

func (mock mockDriverResult) Update(what string, data interface{}) (interface{}, error) {
	*mock.update = true
	return nil, nil
}

type mockDoc struct {
	RecordID int
	IsOut    bool
}

func (doc mockDoc) Table() Table { return "" }
func (doc mockDoc) Id() Thing    { return "" }

func TestDBSelect_Do(t *testing.T) {
	t.Run("mock driver result no error", func(t *testing.T) {
		results, err := SelectOn[mockDoc](NewQueryFrom(Table("")), newMockDriver()).Do()
		require.NoError(t, err)
		assert.Equal(t, []mockDoc{
			{RecordID: 0, IsOut: false},
		}, results)
	})
}

func TestDBSelectAndUpdate_Do(t *testing.T) {
	t.Run("mock driver update no error", func(t *testing.T) {
		db := newMockDriver()
		_, err := SelectAndUpdate(NewQueryFrom(Table("")), func(mockDoc) mockDoc { return mockDoc{} }, db).Do()
		require.NoError(t, err)
		docs, err := SelectOn[mockDoc](NewQueryFrom(Table("")), db).Do()
		require.NoError(t, err)
		assert.Equal(t, []mockDoc{
			{RecordID: 0, IsOut: true},
		}, docs)
	})
}

func (mock mockDriverResult) Create(thing string, data interface{}) (interface{}, error) {
	tb, th := Table("mock"), Thing("mock:00000000_0000_0000_0000_000000000000")
	id, err := NewIDFromThing(th, tb)
	if err != nil {
		return nil, fmt.Errorf("new id from thing: %w", err)
	}
	return struct {
		Id string `json:"id"`
	}{id.String()}, nil
}