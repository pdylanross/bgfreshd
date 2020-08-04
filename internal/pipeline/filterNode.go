package pipeline

import (
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/filter"
	"github.com/sirupsen/logrus"
)

type filterNode struct {
	next   *filterNode
	filter filter.Filter
	logger *logrus.Entry
}

func (fn *filterNode) isValid(img background.Background) bool {
	currentValid := fn.filter.IsValid(img)
	if !currentValid {
		fn.logger.Debug("filter rejected")
		return false
	} else {
		fn.logger.Debug("filter approved")
	}

	return fn.next == nil || fn.next.isValid(img)
}

func (fn *filterNode) copy() *filterNode {
	var next *filterNode = nil
	if fn.next != nil {
		next = fn.next.copy()
	}

	return &filterNode{
		next:   next,
		filter: fn.filter,
		logger: fn.logger,
	}
}

func (fn *filterNode) getLastNode() *filterNode {
	if fn.next != nil {
		return fn.next.getLastNode()
	}

	return fn
}
