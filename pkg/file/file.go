package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Write(dir, file, content string) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dir, file), []byte(fmt.Sprintln(strings.TrimSpace(content))), 0666)
}
