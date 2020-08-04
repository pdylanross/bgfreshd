package pipeline

import (
	"bgfreshd/pkg"
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/source"
)

type sourceNode struct {
	source  source.Source
	filters *filterNode
	weight  uint
}

func (sn *sourceNode) next() (background.Background, error) {
	for true {
		current, err := sn.source.Next()
		if err != nil {
			return nil, err
		}

		if current == nil {
			return nil, &pkg.SourceEmptyError{Source: sn.source}
		}

		found := sn.filters.isValid(current)
		if found {
			return current, nil
		}
	}

	return nil, &pkg.SourceEmptyError{Source: sn.source}
}
