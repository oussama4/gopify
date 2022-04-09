package gopify

import (
	"errors"
	"testing"
)

func TestExtractPagination(t *testing.T) {
	cases := []struct {
		linkHeader         string
		expectedPagination *Pagination
		expectedError      error
	}{
		{
			linkHeader:         "invalid header",
			expectedPagination: nil,
			expectedError:      errors.New("invalid header"),
		},
		{
			linkHeader:         `<https://resource.url?page_info=next_cursor>; rel="next"`,
			expectedPagination: &Pagination{Next: "next_cursor"},
			expectedError:      nil,
		},
		{
			linkHeader:         `<https://resource.url?page_info=next_cursor>; rel="next", <http://resource.url?page_info=previous_cursor>; rel="previous"`,
			expectedPagination: &Pagination{Next: "next_cursor", Previous: "previous_cursor"},
			expectedError:      nil,
		},
	}

	for _, c := range cases {
		pagination, err := extractPagination(c.linkHeader)
		if pagination != c.expectedPagination && c.expectedError != err {
			t.Errorf("expected pagination: %v, and error : %v, but got : %v, %v", c.expectedPagination, c.expectedError, pagination, err)
		}
	}
}
