package config

import (
	"bgfreshd/pkg/filter"
	"bgfreshd/pkg/source"
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"io/ioutil"
)

// Config is the main configuration format
type Config struct {
	OutputPath     string                 `yaml:"outputPath"`
	MaxBackgrounds int                    `yaml:"maxBackgrounds"`
	MinRotationAge int                    `yaml:"minRotationAge"`
	MaxRotationAge int                    `yaml:"maxRotationAge"`
	Sources        []source.Configuration `yaml:"sources"`
	Filters        []filter.Configuration `yaml:"filters"`
}

// Load loads the config at the given path
func Load(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = yaml.Unmarshal(bytes, &config); err != nil {
		fmt.Printf("%v\n", yaml.FormatError(err, true, true))
		return nil, errors.New("error parsing yaml")
	}

	if err = validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation error: %s", err.Error())
	}

	return &config, nil
}

func validate(config *Config) error {
	if len(config.OutputPath) == 0 {
		return errors.New("OutputPath must not be empty")
	}

	for i, _ := range config.Filters {
		if err := validateFilter(&config.Filters[i]); err != nil {
			return err
		}
	}

	for i, _ := range config.Sources {
		if err := validateSource(&config.Sources[i]); err != nil {
			return err
		}
	}

	return nil
}

func validateFilter(config *filter.Configuration) error {
	return nil
}

func validateSource(config *source.Configuration) error {

	if config.Weight == nil {
		config.Weight = new(uint)
		*config.Weight = 1
	}

	return nil
}
