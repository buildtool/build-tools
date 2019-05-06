package ci

import (
  "github.com/stretchr/testify/assert"
  "os"
  "testing"
)

func TestIdentify(t *testing.T) {
  os.Clearenv()

  result := Identify()
  assert.Nil(t, result)
}
