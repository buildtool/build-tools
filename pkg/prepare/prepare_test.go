package prepare

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg/version"
)

func TestDoPrepare(t *testing.T) {
	tests := []struct {
		name   string
		config string
		args   []string
		want   int
	}{
		{
			name: "broken config",
			config: `ci: []
`,
			args: []string{"dummy"},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldPwd, _ := os.Getwd()
			name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
			defer func() { _ = os.RemoveAll(name) }()

			_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(tt.config), 0777)

			err := os.Chdir(name)
			assert.NoError(t, err)
			defer func() { _ = os.Chdir(oldPwd) }()

			if got := DoPrepare(name, version.Info{}, append([]string{"prepare"}, tt.args...)...); got != tt.want {
				t.Errorf("DoPrepare() = %v, want %v", got, tt.want)
			}
		})
	}
}
