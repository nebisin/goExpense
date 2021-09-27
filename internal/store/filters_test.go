package store

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFilters(t *testing.T) {
	type testCases struct {
		column    string
		direction string
		offset    int
		filters   Filters
	}

	cases := []testCases{
		{
			column:    "id",
			direction: "DESC",
			offset:    12,
			filters: Filters{
				Page:  3,
				Limit: 6,
				Sort:  "-id",
			},
		},
		{
			column:    "title",
			direction: "ASC",
			offset:    0,
			filters: Filters{
				Page:  1,
				Limit: 10,
				Sort:  "title",
			},
		},
	}

	for i, value := range cases {
		t.Run(fmt.Sprintf("filters test case #%d", i), func(t *testing.T) {

			column := value.filters.sortColumn()
			require.Equal(t, column, value.column)

			direction := value.filters.sortDirection()
			require.Equal(t, direction, value.direction)

			offset := value.filters.offset()
			require.Equal(t, offset, value.offset)
		})
	}

}
