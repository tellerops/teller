package utils

import (
	"os"
	"path/filepath"
)

const (
	filePermission = 0600
)

func WriteFileInPath(filename, to string, content []byte) error {
	if to != "" {
		if _, err := os.Stat(to); os.IsNotExist(err) {
			err = os.MkdirAll(to, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	err := os.WriteFile(filepath.Join(to, filename), content, filePermission)

	if err != nil {
		return err
	}

	return nil
}
