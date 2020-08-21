package pkg

import (
	"bgfreshd/pkg/source"
	"fmt"
)

type SourceEmptyError struct {
	Source source.Source
}

func (s SourceEmptyError) Error() string {
	return fmt.Sprintf("source %s empty", s.Source.GetName())
}

type NotFoundError struct {
	Key string
}

func (n NotFoundError) Error() string {
	return fmt.Sprintf("item not found: %s", n.Key)
}

