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
