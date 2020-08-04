package pipeline

import (
	"bgfreshd/internal/config"
	"bgfreshd/pkg/background"
	"github.com/mroth/weightedrand"
	"github.com/sirupsen/logrus"
)

// Pipeline is the encapsulation of all of the loading and filtering logic
type Pipeline interface {
	LoadOne() (background.Background, error)
}

// NewPipeline creates the pipeline from config
func NewPipeline(config *config.Config, log *logrus.Logger) (Pipeline, error) {
	pipelineLog := log.WithFields(logrus.Fields{
		"section": "pipeline",
	})
	filterLog := log.WithFields(logrus.Fields{
		"section": "filter",
	})
	sourceLog := log.WithFields(logrus.Fields{
		"section": "source",
	})

	pipelineLog.Info("Building image gathering pipeline")

	globalFilter, err := loadFilters(config.Filters, filterLog)
	if err != nil {
		return nil, err
	}

	sources := make([]sourceNode, 0, len(config.Sources))

	for _, currentSource := range config.Sources {
		sourceFilter, err := loadFilters(currentSource.Filters, filterLog)
		if err != nil {
			return nil, err
		}

		globalCopy := globalFilter.copy()
		last := globalCopy.getLastNode()
		last.next = sourceFilter

		loadedSource, err := CreateSource(&currentSource, sourceLog)
		if err != nil {
			return nil, err
		}

		weight := uint(1)
		if currentSource.Weight != nil {
			weight = *currentSource.Weight
		}

		node := sourceNode{
			source:  loadedSource,
			filters: globalCopy,
			weight:  weight,
		}
		sources = append(sources, node)
	}

	// weightedrand needs rand to be seeded - this happens in cli
	choices := make([]weightedrand.Choice, len(sources))
	for _, s := range sources {
		choices = append(choices, weightedrand.Choice{Item: s, Weight: s.weight})
	}
	sourceChooser := weightedrand.NewChooser(choices...)

	pipelineLog.Info("Pipeline constructed")

	return &pipeline{
		sourceChooser: sourceChooser,
		sources:       sources,
		pipelineLog:   pipelineLog,
		mainLog:       log,
	}, nil
}

func loadFilters(configured []config.BgFilter, filterLog *logrus.Entry) (*filterNode, error) {
	if len(configured) == 0 {
		return nil, nil
	}

	headFilter, err := loadSingleFilter(&configured[0], filterLog)
	if err != nil {
		return nil, err
	}

	if len(configured) == 1 {
		return headFilter, nil
	}

	current := headFilter
	for _, c := range configured[1:] {
		next, err := loadSingleFilter(&c, filterLog)
		if err != nil {
			return nil, err
		}

		current.next = next
		current = next
	}

	return headFilter, nil
}

func loadSingleFilter(conf *config.BgFilter, filterLog *logrus.Entry) (*filterNode, error) {
	filterImpl, err := CreateFilter(conf, filterLog)
	if err != nil {
		return nil, err
	}

	return &filterNode{
		next:   nil,
		filter: filterImpl,
		logger: filterLog.WithFields(logrus.Fields{
			"filter": conf.Type,
		}),
	}, nil
}

type pipeline struct {
	sourceChooser weightedrand.Chooser
	sources       []sourceNode
	pipelineLog   *logrus.Entry
	mainLog       *logrus.Logger
}

func (p pipeline) LoadOne() (background.Background, error) {
	attempts := 0
	p.pipelineLog.Info("Loading image")
	for attempts < 10 {
		attempts++

		source := p.sourceChooser.Pick().(sourceNode)
		bg, err := source.next()
		if err != nil {
			p.pipelineLog.Warnf("error in pipeline process: %s, retrying", err.Error())
		} else {
			return bg, nil
		}

		// retry all errors for now.
		// todo: determine if we have any unrecoverable errors here that shouldn't be retried on
	}

	p.pipelineLog.Error("Pipeline exceeded max retries")
	return nil, &ExceededRetriesError{}
}

type ExceededRetriesError struct {
}

func (p ExceededRetriesError) Error() string {
	return "Pipeline exceeded max retries"
}
