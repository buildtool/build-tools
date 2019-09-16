package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
)

func ExtractFileContent(tarFile io.Reader, filename string) (string, error) {
	r := tar.NewReader(tarFile)

	for {
		header, err := r.Next()
		switch {
		case err == io.EOF:
			return "", fmt.Errorf("file '%s' not found in archive", filename)
		case err != nil:
			return "", err
		}

		fmt.Println(header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if header.Name != filename {
				continue
			}
			buff, err := ioutil.ReadAll(r)
			return string(buff), err
		}
	}
}
