package networking

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPortAvailable(t *testing.T) {
	got := IsPortAvailable(123)
	assert.True(t, got)
}

func TestIsPortAvaibleFail(t *testing.T) {
	port := 80

	// Block port for test
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	require.NoError(t, err)
	defer ln.Close()

	got := IsPortAvailable(port)
	assert.False(t, got)
}

func TestFindRandomPort(t *testing.T) {
	port, err := FindRandomPort()
	require.NoError(t, err)
	require.NotZero(t, port)
}

func TestFindRandomPortInRange(t *testing.T) {
	port, err := FindRandomPortInRange(1, 100)
	require.NoError(t, err)
	require.NotZero(t, port)
}
