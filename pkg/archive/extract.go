package archive

import (
	"HyLauncher/pkg/fileutil"
	"archive/zip"
	"fmt"
	"io"
)

func IsZipValid(path string) error {
	if ok := fileutil.FileExists(path); ok == false {
		return fmt.Errorf("can not find zip: %s", path)
	}

	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("invalid zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open file %s in zip: %w", f.Name, err)
		}

		_, err = io.CopyN(io.Discard, rc, 1)
		rc.Close()

		if err != nil && err != io.EOF {
			return fmt.Errorf("corrupted file %s in zip: %w", f.Name, err)
		}
	}

	return nil
}
