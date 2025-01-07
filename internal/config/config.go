package config

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const configFilename = "inpxer.toml"

type MyConfig struct {
	Storage          string       `toml:"storage"`
	Language         string       `toml:"language"`
	Title            string       `toml:"title"`
	AuthorNameFormat string       `toml:"author_name_format"`
	IndexPath        string       `toml:"index_path"`
	LibraryPath      string       `toml:"library_path"`
	Listen           string       `toml:"listen"`
	FullUrl          string       `toml:"full_url"`
	Converters       []*Converter `toml:"converters"`
}

type Converter struct {
	From      string `toml:"from"`
	To        string `toml:"to"`
	Command   string `toml:"command"`
	Arguments string `toml:"arguments"`
}

func Load() (*MyConfig, error) {
	configFiles := []string{
		configFilename,
		path.Join("/data", configFilename),
	}

	exe, err := os.Executable()
	if err == nil {
		exePath := filepath.Dir(exe)
		exeConf := filepath.Join(exePath, configFilename)
		configFiles = append(configFiles, exeConf)
	}

	var configFile string
	for _, file := range configFiles {
		if _, err := os.Stat(file); err == nil {
			configFile = file
			break
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg MyConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
