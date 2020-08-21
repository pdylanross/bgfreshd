package pipeline

import (
	"bgfreshd/internal/config"
	"bgfreshd/pkg/filter"
	"bgfreshd/pkg/source"
	"fmt"
	"github.com/sirupsen/logrus"
)

// FilterFactoryFunc describes how to construct a filter
type FilterFactoryFunc func(bgFilter *config.BgFilter, log *logrus.Entry) (filter.Filter, error)
// SourceFactoryFunc describes how to construct a source
type SourceFactoryFunc func(cfg *config.BgSource, log *logrus.Entry) (source.Source, error)

var filterRegistrations map[string]FilterFactoryFunc
var sourceRegistrations map[string]SourceFactoryFunc

func AddFilterRegistration(name string, factoryFunc FilterFactoryFunc) {
	if filterRegistrations == nil {
		filterRegistrations = make(map[string]FilterFactoryFunc)
	}
	filterRegistrations[name] = factoryFunc
}

func AddSourceRegistration(name string, factoryFunc SourceFactoryFunc) {
	if sourceRegistrations == nil {
		sourceRegistrations = make(map[string]SourceFactoryFunc)
	}

	sourceRegistrations[name] = factoryFunc
}

type FilterNotFoundError struct {
	filterName string
}

func (f *FilterNotFoundError) Error() string {
	return fmt.Sprintf("filter %s not found", f.filterName)
}

func CreateFilter(config *config.BgFilter, filterLog *logrus.Entry) (filter.Filter, error) {
	filterLog.Infof("Constructing filter of type %s", config.Type)
	individualLog := filterLog.WithFields(logrus.Fields{
		"filter": config.Type,
	})

	factory, ok := filterRegistrations[config.Type]
	if !ok {
		filterLog.Errorf("Could not find filter of type %s", config.Type)
		return nil, &FilterNotFoundError{filterName: config.Type}
	}

	return factory(config, individualLog)
}

type SourceNotFoundError struct {
	sourceName string
}

func (s *SourceNotFoundError) Error() string {
	return fmt.Sprintf("source %s not found", s.sourceName)
}

func CreateSource(config *config.BgSource, sourceLog *logrus.Entry) (source.Source, error) {
	sourceLog.Infof("Constructing source of type %s", config.Type)
	individualLog := sourceLog.WithFields(logrus.Fields{
		"source": config.Type,
	})

	factory, ok := sourceRegistrations[config.Type]
	if !ok {
		sourceLog.Errorf("Could not find source of type %s", config.Type)
		return nil, &SourceNotFoundError{sourceName: config.Type}
	}

	return factory(config, individualLog)
}
