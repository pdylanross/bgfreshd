package source

import (
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/filter"
	"github.com/sirupsen/logrus"
	"time"
)

// FactoryFunc describes how to construct a source
type FactoryFunc func(cfg *Configuration, dbFactory DbFactoryFunc, log *logrus.Entry) (Source, error)

// DbFactoryFunc creates a source's db based on a source's metadata
// this helps the source construction pipeline determine which db subset to provide for the source
// this also can optimize away db usage in the case that a specific source impl doesn't need a db
//
// note: the metadata should be immutable based on the configuration of the source.
type DbFactoryFunc func(meta string) (Db, error)

// Source gets background candidates from some location or feed
type Source interface {
	Next() (background.Background, error)
	GetName() string
}

// SourceConfiguration describes settings and configuration of a source
type Configuration struct {
	// Filters that apply to this specific source
	Filters []filter.Configuration `yaml:"filters"`

	// Type of the source
	Type string `yaml:"type"`

	// Options specific to the source
	Options interface{} `yaml:"options"`

	// Weight granted to the source
	Weight *uint `yaml:"weight,omitempty"`
}

// Db describes how sources track relevant progress information
type Db interface {
	SetString(key string, val string) error
	GetString(key string) (string, error)
	SetTime(key string, time time.Time) error
	GetTime(key string) (time.Time, error)
	SetBool(key string, val bool) error
	GetBool(key string) (bool, error)
	KeyExists(key string) (bool, error)
}
