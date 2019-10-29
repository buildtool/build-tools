package file

import (
	"fmt"
	"github.com/buildtool/build-tools/pkg/templating"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var name string

func TestMain(m *testing.M) {
	tempDir := setup()
	code := m.Run()
	teardown(tempDir)
	os.Exit(code)
}

func setup() string {
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

func TestAppend_Return_Error_For_Missing_File(t *testing.T) {
	err := Append(filepath.Join(name, "missing_dir/missing_file_XYZ"), "content")
	assert.EqualError(t, err, fmt.Sprintf("open %s/missing_dir/missing_file_XYZ: no such file or directory", name))
}

func TestAppend_Appends_To_Existing_File(t *testing.T) {
	fileName := filepath.Join(name, "file")
	defer func() { _ = os.RemoveAll(fileName) }()
	err := ioutil.WriteFile(fileName, []byte("abc"), 0777)
	assert.NoError(t, err)
	err = Append(fileName, "content")
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, "abc\ncontent\n", string(bytes))
}

func TestAppendTemplated_Return_Error_For_Broken_Template(t *testing.T) {
	fileName := filepath.Join(name, "file")
	defer func() { _ = os.RemoveAll(fileName) }()
	err := ioutil.WriteFile(fileName, []byte("abc"), 0777)
	assert.NoError(t, err)
	err = AppendTemplated(fileName, "--->{{.ProjectName }<---", templating.TemplateData{ProjectName: "ABC"})
	assert.EqualError(t, err, `template: content:1: unexpected "}" in operand`)
}

func TestAppendTemplated_Appends_To_Existing_File(t *testing.T) {
	fileName := filepath.Join(name, "file")
	defer func() { _ = os.RemoveAll(fileName) }()
	err := ioutil.WriteFile(fileName, []byte("abc"), 0777)
	assert.NoError(t, err)
	err = AppendTemplated(fileName, "--->{{.ProjectName }}<---", templating.TemplateData{ProjectName: "ABC"})
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, "abc\n--->ABC<---\n", string(bytes))
}

func TestWrite_Creates_All_Parent_Directories(t *testing.T) {
	fileName := filepath.Join(name, "missing", "path", "file")
	defer func() { _ = os.RemoveAll(fileName) }()
	err := Write(name, filepath.Join("missing", "path", "file"), "abc")
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, "abc\n", string(bytes))
}

func TestWriteTemplated_Return_Error_For_Broken_Template(t *testing.T) {
	fileName := filepath.Join(name, "missing", "path", "file")
	defer func() { _ = os.RemoveAll(fileName) }()
	err := WriteTemplated(name, filepath.Join("missing", "path", "file"), "{{ .ProjectName }", templating.TemplateData{ProjectName: "ABC"})
	assert.EqualError(t, err, `template: content:1: unexpected "}" in operand`)
}

func TestWriteTemplated(t *testing.T) {
	fileName := filepath.Join(name, "missing", "path", "file")
	defer func() { _ = os.RemoveAll(fileName) }()
	err := WriteTemplated(name, filepath.Join("missing", "path", "file"), "--->{{ .ProjectName }}<---", templating.TemplateData{ProjectName: "ABC"})
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, "--->ABC<---\n", string(bytes))
}
