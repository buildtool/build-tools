package file

import (
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg/templating"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Append(name, content string) error {
	if f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return err
	} else {
		defer func() { _ = f.Close() }()
		_, err := f.WriteString(fmt.Sprintf("\n%s\n", content))
		return err
	}
}

func AppendTemplated(name, template string, data templating.TemplateData) error {
	if content, err := templating.Execute(template, data); err != nil {
		return err
	} else {
		return Append(name, content)
	}
}

func Write(dir, file, content string) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dir, file), []byte(fmt.Sprintln(strings.TrimSpace(content))), 0666)
}

func WriteTemplated(dir, file, template string, data templating.TemplateData) error {
	if content, err := templating.Execute(template, data); err != nil {
		return err
	} else {
		return Write(dir, file, content)
	}
}
