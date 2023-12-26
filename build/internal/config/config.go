package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	SiteTitle       string `json:"siteTitle"`
	SiteDescription string `json:"siteDescription"`
	TemplatePath    string `json:"templatePath"`
	ContentPath     string `json:"contentPath"`
	OutputPath      string `json:"outputPath"`
	ThemeName       string `json:"themeName"`
	DataPath        string `json:"dataPath"`
	// Add other configuration fields as needed
}

func LoadConfig(path string) (*Config, error) {
	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
