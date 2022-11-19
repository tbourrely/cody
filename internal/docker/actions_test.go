package docker

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateName(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple",
			path:     filepath.Join("home", "test", "directory"),
			expected: "directory",
		},
		{
			name:     "spaces",
			path:     filepath.Join("home", "user", "firstname lastname"),
			expected: "firstname_lastname",
		},
		{
			name:     "symbols",
			path:     filepath.Join("home", "test", "41\\name$("),
			expected: "41_name__",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := containerName(tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
