package source

import (
	"bgfreshd/pkg/background"
)

type Source interface {
	Next() (background.Background, error)
	GetName() string
}
