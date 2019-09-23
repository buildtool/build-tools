package tar

import (
	"archive/tar"
	"bytes"
	"errors"
	"github.com/docker/docker/pkg/archive"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"testing"
)

func TestExtractFileContent_Missing_File_Return_Error(t *testing.T) {
	file, _ := archive.Generate("OtherFile", "abc")

	_, err := ExtractFileContent(file, "Dockerfile")
	assert.EqualError(t, err, "file 'Dockerfile' not found in archive")
}

func TestExtractFileContent_Broken_Archive(t *testing.T) {
	_, err := ExtractFileContent(&brokenReader{}, "Dockerfile")

	assert.EqualError(t, err, "read error")
}

func TestExtractFileContent_Directories_Are_Ignored(t *testing.T) {
	buff := &bytes.Buffer{}
	w := tar.NewWriter(buff)
	_ = w.WriteHeader(&tar.Header{Name: "test", Typeflag: tar.TypeDir})
	_, err := ExtractFileContent(buff, "Dockerfile")

	assert.EqualError(t, err, "file 'Dockerfile' not found in archive")
}

func TestExtractFileContent(t *testing.T) {
	file, _ := archive.Generate("Dockerfile", "abc")

	result, err := ExtractFileContent(file, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, "abc", result)
}

func TestExtractFileContentError(t *testing.T) {
	file, _ := archive.Generate("Dockerfile", "abc")
	data, err := ioutil.ReadAll(file)
	assert.NoError(t, err)

	buff := bytes.NewBuffer(data)
	reader := &brokenTarAwareReader{buff: buff}
	_, err = ExtractFileContent(reader, "Dockerfile")
	assert.EqualError(t, err, "read error")
}

type brokenReader struct{}

func (b brokenReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ io.Reader = &brokenReader{}

type brokenTarAwareReader struct {
	buff  *bytes.Buffer
	calls int
}

func (f *brokenTarAwareReader) Read(p []byte) (n int, err error) {
	if f.calls == 1 {
		return 0, errors.New("read error")
	}
	f.calls = f.calls + 1
	i, err := f.buff.Read(p)
	return i, err
}

var _ io.Reader = &brokenTarAwareReader{}
