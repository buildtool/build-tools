package ci

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestIdentify(t *testing.T) {
	os.Clearenv()

	_, err := Identify()
	assert.EqualError(t, err, "no CI found")
}
