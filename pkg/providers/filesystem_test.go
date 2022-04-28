package providers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
)

func createMockDirectoryStructure(f *FileSystem) error {
	createFileSystemData := []struct {
		path     string
		fileName string
		value    string
	}{
		{"settings/prod", "billing-svc", "shazam"},
		{"settings/prod/billing/all", "secret-a", "mailman"},
		{"settings/prod/billing/all", "secret-b", "shazam"},
		{"settings/prod/billing/all/folder", "secret-c", "shazam-1"},
	}

	for _, filePath := range createFileSystemData {
		err := os.MkdirAll(filepath.Join(f.rootDirectory, filePath.path), os.ModePerm)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(f.rootDirectory, filePath.path, filePath.fileName), []byte(filePath.value), 0644)
		if err != nil {
			return err
		}

	}
	return nil
}

func TestFileSystem(t *testing.T) {

	tempFolder, err := os.MkdirTemp(os.TempDir(), "teller-filesystem")
	assert.Nil(t, err)
	defer os.RemoveAll(tempFolder)

	f := &FileSystem{
		logger:        GetTestLogger(),
		rootDirectory: tempFolder,
	}

	err = createMockDirectoryStructure(f)
	assert.NoError(t, err)

	AssertProvider(t, f, false)
	ents, err := f.GetMapping(core.KeyPath{Path: "settings/prod/billing/all", Decrypt: true})
	assert.Nil(t, err)
	assert.Equal(t, len(ents), 3)
}

func TestFileSystemSetEntry(t *testing.T) {

	tempFolder, err := os.MkdirTemp(os.TempDir(), "teller-filesystem")
	assert.Nil(t, err)
	defer os.RemoveAll(tempFolder)

	f := &FileSystem{
		logger:        GetTestLogger(),
		rootDirectory: tempFolder,
	}
	err = createMockDirectoryStructure(f)
	assert.NoError(t, err)

	destFile := "create/newfolder/foo"
	_, err = f.Get(core.KeyPath{Path: destFile, Decrypt: true})
	assert.NotEmpty(t, err)

	err = f.Put(core.KeyPath{Path: destFile}, "new-val")
	assert.Nil(t, err)

	results, err := f.Get(core.KeyPath{Path: destFile, Decrypt: true})
	assert.Nil(t, err)
	assert.NotEmpty(t, results)

}

func TestFileSystemDeleteEntry(t *testing.T) {

	tempFolder, err := os.MkdirTemp(os.TempDir(), "teller-filesystem")
	assert.Nil(t, err)
	defer os.RemoveAll(tempFolder)

	f := &FileSystem{
		logger:        GetTestLogger(),
		rootDirectory: tempFolder,
	}
	err = createMockDirectoryStructure(f)
	assert.NoError(t, err)

	destFile := "settings/prod/billing-svc"
	_, err = f.Get(core.KeyPath{Path: destFile, Decrypt: true})
	assert.Nil(t, err)

	err = f.Delete(core.KeyPath{Path: destFile, Decrypt: true})
	assert.Nil(t, err)

	_, err = f.Get(core.KeyPath{Path: destFile, Decrypt: true})
	assert.NotNil(t, err)

}
