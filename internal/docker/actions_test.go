package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_safeContainerName(t *testing.T) {
	testCases := []struct {
		name     string
		dirName  string
		expected string
	}{
		{
			name:     "simple",
			dirName:  "directory",
			expected: "directory",
		},
		{
			name:     "spaces",
			dirName:  "firstname lastname",
			expected: "firstname_lastname",
		},
		{
			name:     "symbols",
			dirName:  "41\\name$(",
			expected: "41_name__",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := safeContainerName(tc.dirName)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
