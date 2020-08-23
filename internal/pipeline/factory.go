package pipeline

import (
	"bgfreshd/internal/db"
	"bgfreshd/pkg/filter"
	"bgfreshd/pkg/source"
	"fmt"
	"github.com/sirupsen/logrus"
)

var filterRegistrations map[string]filter.FactoryFunc
var sourceRegistrations map[string]source.FactoryFunc

func AddFilterRegistration(name string, factoryFunc filter.FactoryFunc) {
	if filterRegistrations == nil {
		filterRegistrations = make(map[string]filter.FactoryFunc)
	}
	filterRegistrations[name] = factoryFunc
}

func AddSourceRegistration(name string, factoryFunc source.FactoryFunc) {
	if sourceRegistrations == nil {
		sourceRegistrations = make(map[string]source.FactoryFunc)
	}

	sourceRegistrations[name] = factoryFunc
}

type FilterNotFoundError struct {
	filterName string
}

func (f *FilterNotFoundError) Error() string {
	return fmt.Sprintf("filter %s not found", f.filterName)
}

func CreateFilter(config *filter.Configuration, filterLog *logrus.Entry) (filter.Filter, error) {
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

func CreateSource(config *source.Configuration, db db.BackgroundDb, sourceLog *logrus.Entry) (source.Source, error) {
	sourceLog.Infof("Constructing source of type %s", config.Type)
	individualLog := sourceLog.WithFields(logrus.Fields{
		"source": config.Type,
	})

	factory, ok := sourceRegistrations[config.Type]
	if !ok {
		sourceLog.Errorf("Could not find source of type %s", config.Type)
		return nil, &SourceNotFoundError{sourceName: config.Type}
	}

	return factory(
		config,
		func(meta string) (source.Db, error) {
			return db.NewSourceDb(config.Type, meta)
		},
		individualLog)
}
