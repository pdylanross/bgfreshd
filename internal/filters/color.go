package filters

import (
	"bgfreshd/internal"
	"bgfreshd/internal/config"
	"bgfreshd/internal/pipeline"
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/filter"
	"github.com/sirupsen/logrus"
	"image"
	"image/color"
	"math"
)

type ColorOptions struct {
	DesiredColor       *ColorRGBValue `yaml:"desiredColor"`
	AcceptableDistance float64        `yaml:"acceptableDistance"`
}

type ColorRGBValue struct {
	Red   uint8 `yaml:"red"`
	Green uint8 `yaml:"green"`
	Blue  uint8 `yaml:"blue"`
}

func init() {
	pipeline.AddFilterRegistration("color", NewColorFilter)
}

func NewColorFilter(config *config.BgFilter, filterLog *logrus.Entry) (filter.Filter, error) {
	var options ColorOptions
	if err := internal.CastDecodedYamlToType(config.Options, &options); err != nil {
		return nil, err
	}

	return &colorFilter{
		filterLog: filterLog,
		opt:       &options,
	}, nil
}

type colorFilter struct {
	filterLog *logrus.Entry
	opt       *ColorOptions
}

func (c *colorFilter) IsValid(img background.Background) bool {
	avg := computeAverageColor(img.GetImage())

	c.filterLog.Debugf("avg r: %d g: %d b: %d", avg.R, avg.G, avg.B)
	dist := computeDistance(c.opt.DesiredColor, avg)
	c.filterLog.Debugf("distance: %2f", dist)
	return dist < c.opt.AcceptableDistance
}

// there exist methods to compute this better
// take the simple route for now
func computeAverageColor(img image.Image) color.NRGBA {
	bounds := img.Bounds()
	points := uint64(0)
	red := uint64(0)
	green := uint64(0)
	blue := uint64(0)
	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			curColor := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			points++
			red += uint64(curColor.R)
			green += uint64(curColor.G)
			blue += uint64(curColor.B)
		}
	}

	avgRed := red / points
	avgGreen := green / points
	avgBlue := blue / points

	return color.NRGBA{
		R: uint8(avgRed),
		G: uint8(avgGreen),
		B: uint8(avgBlue),
		A: 255,
	}
}

// simple euclidean distance for now
// there are better methods around
func computeDistance(a *ColorRGBValue, b color.NRGBA) float64 {
	ar := a.Red
	ag := a.Green
	ab := a.Blue
	br := b.R
	bg := b.G
	bb := b.B

	dR := (br - ar) * (br - ar)
	dG := (bg - ag) * (bg - ag)
	dB := (bb - ab) * (bb - ab)
	distance := math.Sqrt(float64(dR + dB + dG))

	dC := (255) ^ 2
	maxColorDistance := math.Sqrt(float64(dC + dC + dC))
	return distance / maxColorDistance
}
