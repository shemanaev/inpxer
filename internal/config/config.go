package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type MyConfig struct {
	Language    string       `toml:"language"`
	Title       string       `toml:"title"`
	IndexPath   string       `toml:"index_path"`
	LibraryPath string       `toml:"library_path"`
	Listen      string       `toml:"listen"`
	FullUrl     string       `toml:"full_url"`
	Converters  []*Converter `toml:"converters"`
}

type Converter struct {
	From      string `toml:"from"`
	To        string `toml:"to"`
	Command   string `toml:"command"`
	Arguments string `toml:"arguments"`
}

func Load() (*MyConfig, error) {
	configFiles := []string{
		"inpxer.toml",
		"/data/inpxer.toml",
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
