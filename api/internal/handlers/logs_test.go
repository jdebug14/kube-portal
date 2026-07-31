package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTailLines(t *testing.T) {
	tests := []struct {
		caseName  string
		raw       string
		expected  int64
		expectErr bool
	}{
		{caseName: "empty defaults to 100", raw: "", expected: 100},
		{caseName: "valid value passes through", raw: "250", expected: 250},
		{caseName: "over cap clamps to 1000", raw: "5000", expected: 1000},
		{caseName: "exactly at cap", raw: "1000", expected: 1000},
		{caseName: "invalid string", raw: "notanumber", expectErr: true},
		{caseName: "zero", raw: "0", expectErr: true},
		{caseName: "negative", raw: "-5", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.caseName, func(t *testing.T) {
			result, err := parseTailLines(tc.raw)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
