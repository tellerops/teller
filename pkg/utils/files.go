package utils

import (
	"os"
	"path/filepath"
)

func WriteFileInPath(filename string, to string, content []byte) error {
	if (to != "") {
		if _, err := os.Stat(to); os.IsNotExist(err) {
			err = os.MkdirAll(to, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	err := os.WriteFile(filepath.Join(to, filename), content, 0600)

	if (err != nil) {
		return err
	}

	return nil
}