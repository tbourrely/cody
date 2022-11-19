package configuration

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cody/internal/types"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

func generateFromYaml(content []byte) (types.Configuration, error) {
	config := types.Configuration{}

	err := yaml.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func loadByDir(fileSystem fs.FS, dir string) (config types.Configuration, err error) {
	configPath := filepath.Join(dir, types.CONFIGURATION_FILENAME)
	homeConfigContent, err := fs.ReadFile(fileSystem, configPath[1:]) // remove first slash
	if err == nil {
		config, err = generateFromYaml(homeConfigContent)
	}

	return
}

// Load loads config from cody.yml files.
// It will first attempt to load the file from the user home directory,
// and then try to load the configuration from the current directory.
func Load(fileSystem fs.FS) (config types.Configuration, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	homeConfig, homeConfigErr := loadByDir(fileSystem, home)
	if homeConfigErr == nil {
		mergo.MergeWithOverwrite(&config, homeConfig)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	cwdConfig, cwdConfigErr := loadByDir(fileSystem, cwd)
	if cwdConfigErr == nil {
		mergo.MergeWithOverwrite(&config, cwdConfig)
	}

	return
}
