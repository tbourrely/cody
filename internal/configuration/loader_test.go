package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cody/internal/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFromYaml(t *testing.T) {
	var data = `
ports:
  start: 1
  end: 3
`

	config, err := generateFromYaml([]byte(data))

	require.NoError(t, err)

	assert.Equal(t, 1, config.Ports.Start)
	assert.Equal(t, 3, config.Ports.End)
}

func TestLoad(t *testing.T) {
	var data = `
ports:
  start: 10
  end: 30
auth_token: 'test_token_value'
`
	var fs = afero.NewOsFs()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	path := filepath.Join(cwd, types.CONFIGURATION_FILENAME)
	defer fs.Remove(path)
	afero.WriteFile(fs, path, []byte(data), os.ModePerm)

	config, err := Load(os.DirFS("/"))

	require.NoError(t, err)
	assert.Equal(t, 10, config.Ports.Start)
	assert.Equal(t, 30, config.Ports.End)
	assert.Equal(t, "test_token_value", config.AuthToken)
}

func TestLoadHome(t *testing.T) {
	var data = `
ports:
  start: 2
  end: 40
`
	var fs = afero.NewOsFs()

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	path := filepath.Join(home, types.CONFIGURATION_FILENAME)
	defer fs.Remove(path)
	afero.WriteFile(fs, path, []byte(data), os.ModePerm)

	config, err := Load(os.DirFS("/"))

	require.NoError(t, err)
	assert.Equal(t, 2, config.Ports.Start)
	assert.Equal(t, 40, config.Ports.End)
}

func TestLoadMerge(t *testing.T) {
	var dataHome = `
ports:
  start: 1
  end: 40
`
	var dataCwd = `
ports:
  end: 30
`

	var fs = afero.NewOsFs()

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	homePath := filepath.Join(home, types.CONFIGURATION_FILENAME)
	defer fs.Remove(homePath)
	afero.WriteFile(fs, homePath, []byte(dataHome), os.ModePerm)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	cwdPath := filepath.Join(cwd, types.CONFIGURATION_FILENAME)
	defer fs.Remove(cwdPath)
	afero.WriteFile(fs, cwdPath, []byte(dataCwd), os.ModePerm)

	config, err := Load(os.DirFS("/"))

	require.NoError(t, err)
	assert.Equal(t, 1, config.Ports.Start)
	assert.Equal(t, 30, config.Ports.End)
}
