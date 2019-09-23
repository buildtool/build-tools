package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
)

func ExtractFileContent(tarFile io.Reader, filename string) (string, error) {
	r := tar.NewReader(tarFile)
	var content *string
	for {
		header, err := r.Next()
		switch {
		case err == io.EOF:
			if content == nil {
				return "", fmt.Errorf("file '%s' not found in archive", filename)
			} else {
				return *content, nil
			}
		case err != nil:
			return "", err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if header.Name != filename {
				continue
			}
			buff, err := ioutil.ReadAll(r)
			if err != nil {
				return "", err
			}
			s := string(buff)
			content = &s
		}
	}
}
