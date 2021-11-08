package ci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpCI_Name(t *testing.T) {
}

func TestNoOpCI_Override_ImageName(t *testing.T) {
	ci := &No{Common: &Common{}}
	ci.SetImageName("override")
	assert.Equal(t, "override", ci.BuildName())
}
