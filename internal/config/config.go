package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port               int           `yaml:"port"`
	Storage            string        `yaml:"storage"`
	FileTypes          []string      `yaml:"file_types"`
	Timeout            time.Duration `yaml:"timeout"`
	MaxFileSize        int           `yaml:"max_file_size_mb"`
	MaxProcessingTasks int           `yaml:"max_processing_tasks"`
}

func NewConfig(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}
