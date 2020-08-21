package source

import (
	"bgfreshd/pkg/background"
	"time"
)

type Source interface {
	Next() (background.Background, error)
	GetName() string
}

type Db interface {
	SetString(key string, val string) error
	GetString(key string) (string, error)
	SetTime(key string, time time.Time) error
	GetTime(key string) (time.Time, error)
	SetBool(key string, val bool) error
	GetBool(key string) (bool, error)
	KeyExists(key string) (bool, error)
}
