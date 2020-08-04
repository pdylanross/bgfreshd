package pipeline

import (
	"bgfreshd/internal/config"
	"bgfreshd/internal/sources"
	"bgfreshd/pkg/filter"
	"bgfreshd/pkg/source"
	"fmt"
	"github.com/sirupsen/logrus"
)

type FilterFactoryFunc func(bgFilter *config.BgFilter, log *logrus.Entry) (filter.Filter, error)

var filterRegistrations map[string]FilterFactoryFunc

func AddFilterRegistration(name string, factoryFunc FilterFactoryFunc) {
	if filterRegistrations == nil {
		filterRegistrations = make(map[string]FilterFactoryFunc)
	}
	filterRegistrations[name] = factoryFunc
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

	switch config.Type {
	case "reddit":
		return sources.NewRedditSource(config, individualLog)
	}

	sourceLog.Errorf("Could not find source type %s!", config.Type)
	return nil, &SourceNotFoundError{sourceName: config.Type}
}
