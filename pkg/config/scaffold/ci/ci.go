package ci

import (
	"github.com/sparetimecoders/build-tools/pkg/templating"
)

type CI interface {
	Name() string
	ValidateConfig() error
	Validate(name string) error
	Scaffold(dir string, data templating.TemplateData) (*string, error)
	Badges(name string) ([]templating.Badge, error)
	Configure() error
}
