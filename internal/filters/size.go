package filters

import (
	"bgfreshd/internal"
	"bgfreshd/internal/config"
	"bgfreshd/internal/pipeline"
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/filter"
	"github.com/sirupsen/logrus"
	"math"
)

type SizeOptions struct {
	MinX int `yaml:"minX"`
	MinY int `yaml:"minY"`
	MaxX int `yaml:"maxX"`
	MaxY int `yaml:"maxY"`
}

func init() {
	pipeline.AddFilterRegistration("size", NewSizeFilter)
}

func NewSizeFilter(config *config.BgFilter, filterLog *logrus.Entry) (filter.Filter, error) {
	var options SizeOptions
	if err := internal.CastDecodedYamlToType(config.Options, &options); err != nil {
		return nil, err
	}

	// default max vals if unset
	if options.MaxX == 0 {
		options.MaxX = math.MaxInt32
	}
	if options.MaxY == 0 {
		options.MaxY = math.MaxInt32
	}

	return &sizeFilter{
		filterLog: filterLog,
		opts:      &options,
	}, nil
}

type sizeFilter struct {
	filterLog *logrus.Entry
	opts      *SizeOptions
}

func (c *sizeFilter) IsValid(img background.Background) bool {
	i := img.GetImage()
	bounds := i.Bounds()
	c.filterLog.Debugf("image size x: %d y: %d", bounds.Max.X, bounds.Max.Y)

	return bounds.Max.X >= c.opts.MinX && bounds.Max.Y >= c.opts.MinY && bounds.Max.X <= c.opts.MaxX && bounds.Max.Y <= c.opts.MaxY
}
