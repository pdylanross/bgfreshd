package internal

import (
	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
)

// CastDecodedYamlToType makes up for go-yaml not having a similar concept to
// json.RawMessage by marshaling and unmarshaling a generic interface to and from yaml
func CastDecodedYamlToType(in interface{}, out interface{}) error {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, out)
}

func Deferrer(logger *logrus.Entry, fun func() error) {
	if err := fun(); err != nil {
		logger.Warnf("Error in deferred func: %s", err.Error())
	}
}
