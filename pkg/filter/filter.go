package filter

import "bgfreshd/pkg/background"

type Filter interface {
	IsValid(img background.Background) bool
}
