package filter

import (
	"bgfreshd/pkg/background"
	"github.com/sirupsen/logrus"
)

// FilterFactoryFunc describes how to construct a filter
type FactoryFunc func(bgFilter *Configuration, log *logrus.Entry) (Filter, error)

type Filter interface {
	IsValid(img background.Background) bool
}

// Configuration describes settings and type of a filter
type Configuration struct {
	// Type of the filter
	Type string `yaml:"type"`

	// Options specific to the filter
	Options map[string]interface{} `yaml:"options"`
}
