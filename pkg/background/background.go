package background

import (
	"image"
	"math/rand"
	"time"
)

type Background interface {
	GetName() string
	GetCreatedDate() time.Time
	GetMetadataKeys() []string
	GetMetadata(name string) string
	AddMetadata(name string, value string)
	GetImage() image.Image
	IsActive() bool
	SetActive()
	SetInactive()
	GetExpiry() time.Time
	GenerateExpiry(minDays int, maxDays int)
}

func FromImage(img image.Image, identifier string) Background {
	return &bg{
		name:               identifier,
		image:              img,
		createDate:         time.Now(),
		additionalMetadata: map[string]string{},
		isActive:           false,
	}
}

type bg struct {
	name               string
	isActive           bool
	createDate         time.Time
	expiresOnDate      time.Time
	image              image.Image
	additionalMetadata map[string]string
}

func (b *bg) IsActive() bool {
	return b.isActive
}

func (b *bg) SetActive() {
	b.isActive = true
}

func (b *bg) SetInactive() {
	b.isActive = true
}

func (b *bg) GetMetadataKeys() []string {
	var keys []string
	for k := range b.additionalMetadata {
		keys = append(keys, k)
	}
	return keys
}

func (b *bg) GetMetadata(name string) string {
	return b.additionalMetadata[name]
}

func (b *bg) AddMetadata(name string, value string) {
	b.additionalMetadata[name] = value
}

func (b *bg) GetImage() image.Image {
	return b.image
}

func (b *bg) GetName() string {
	return b.name
}

func (b *bg) GetCreatedDate() time.Time {
	return b.createDate
}

func (b *bg) GetExpiry() time.Time {
	return b.expiresOnDate
}

func (b *bg) GenerateExpiry(minDays int, maxDays int) {
	difference := maxDays - minDays
	expiresInDays := minDays + rand.Intn(difference)
	expiresInHours := 24 * expiresInDays
	b.expiresOnDate = b.createDate.Add(time.Hour * time.Duration(expiresInHours))
}
