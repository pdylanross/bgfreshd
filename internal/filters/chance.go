package filters

import (
	"bgfreshd/internal"
	"bgfreshd/internal/config"
	"bgfreshd/internal/pipeline"
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/filter"
	"github.com/sirupsen/logrus"
	"math/rand"
)

type ChanceOptions struct {
	Chance float64 `yaml:"chance"`
}

func init() {
	pipeline.AddFilterRegistration("chance", NewChanceFilter)
}

func NewChanceFilter(config *config.BgFilter, filterLog *logrus.Entry) (filter.Filter, error) {
	var options ChanceOptions
	if err := internal.CastDecodedYamlToType(config.Options, &options); err != nil {
		return nil, err
	}

	return &chanceFilter{
		filterLog: filterLog,
		chance:    options.Chance,
	}, nil
}

type chanceFilter struct {
	filterLog *logrus.Entry
	chance    float64
}

func (c *chanceFilter) IsValid(img background.Background) bool {
	return rand.Float64() < c.chance
}
