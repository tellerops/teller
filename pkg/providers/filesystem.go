package providers

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/karrick/godirwalk"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type FileSystem struct {
	logger        logging.Logger
	rootDirectory string
}

const FileSystemName = "FileSystem"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "File system",
		Name:           FileSystemName,
		Authentication: "",
		ConfigTemplate: `
  filesystem:
    env_sync:
      path: redis/config
    env:
      ETC_DSN:
        path: redis/config/foobar
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true, Delete: true},
	}

	RegisterProvider(metaInfo, NewFileSystem)
}

// NewFileSystem creates new provider instance
func NewFileSystem(logger logging.Logger) (core.Provider, error) {
	return &FileSystem{
		logger:        logger,
		rootDirectory: "",
	}, nil
}

// Put will create a new single entry
func (f *FileSystem) Put(p core.KeyPath, val string) error {
	return f.writeFile(f.getFilePath(p.Path), val)
}

// PutMapping will create a multiple entries
func (f *FileSystem) PutMapping(p core.KeyPath, m map[string]string) error {
	for k, v := range m {
		ap := p.SwitchPath(fmt.Sprintf("%v/%v", f.getFilePath(p.Path), k))
		err := f.Put(ap, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetMapping returns a multiple entries
func (f *FileSystem) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {

	findings := []core.EnvEntry{}
	err := godirwalk.Walk(f.getFilePath(p.Path), &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {

			if !de.IsRegular() || strings.HasPrefix(de.Name(), ".") {
				return nil
			}
			content, err := f.readFile(osPathname)
			if err != nil {
				f.logger.WithError(err).WithField("path", p.Path).Debug("file not found in path")
				return nil
			}

			if !f.IsText(content) {
				return nil
			}
			findings = append(findings, p.FoundWithKey(strings.Replace(path.Clean(osPathname), fmt.Sprintf("%s/", p.Path), "", 1), string(content)))

			return nil
		},
		Unsorted: true,
	})

	return findings, err
}

// Get returns a single entry
func (f *FileSystem) Get(p core.KeyPath) (*core.EnvEntry, error) {
	content, err := f.readFile(f.getFilePath(p.Path))
	if err != nil {
		f.logger.WithError(err).WithField("path", p.Path).Debug("file not found in path")
		return nil, err
	}
	ent := p.Found(string(content))
	return &ent, nil
}

// Delete will delete entry
func (f *FileSystem) Delete(kp core.KeyPath) error {
	deletePath := f.getFilePath(kp.Path)
	fileInfo, err := os.Stat(deletePath)
	if err != nil {
		return err
	}
	// to make the delete safely, we allow deleting a single file only
	if fileInfo.IsDir() {
		return errors.New("delete folder is not supported")
	}

	return os.Remove(deletePath)
}

// DeleteMapping will delete the given path
func (f *FileSystem) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("provider mapping %s does not implement delete yet", FileSystemName)
}

func (f *FileSystem) getFilePath(p string) string {
	if f.rootDirectory == "" {
		return p
	}
	return filepath.Join(f.rootDirectory, p)
}

func (f *FileSystem) writeFile(to, val string) error {
	f.logger.WithField("path", to).Info("put entry value")
	dir, _ := path.Split(to)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		f.logger.WithField("dir", dir).Debug("create folder path")
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(to, []byte(val), 0600) //nolint
}

func (f *FileSystem) readFile(filePath string) ([]byte, error) {

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	content = bytes.TrimSuffix(content, []byte("\n"))
	content = bytes.TrimSuffix(content, []byte("\r\n"))
	return content, nil
}

func (f *FileSystem) IsText(s []byte) bool {
	const max = 1024
	if len(s) > max {
		s = s[0:max]
	}
	for i, c := range string(s) {
		if i+utf8.UTFMax > len(s) {
			break
		}
		if c == 0xFFFD || c < ' ' && c != '\n' && c != '\t' && c != '\f' {
			return false
		}
	}
	return true
}
