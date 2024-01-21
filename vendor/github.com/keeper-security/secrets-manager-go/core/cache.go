package core

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const defautFilePath = "ksm_cache.bin"

type ICache interface {
	SaveCachedValue(data []byte) error
	GetCachedValue() ([]byte, error)
	Purge() error
}

// File based cache
type fileCache struct {
	FilePath string
}

func (c *fileCache) SaveCachedValue(data []byte) error {
	if data == nil {
		data = []byte{}
	}
	return ioutil.WriteFile(c.FilePath, data, 0600)
}

func (c *fileCache) GetCachedValue() ([]byte, error) {
	return ioutil.ReadFile(c.FilePath)
}

func (c *fileCache) Purge() error {
	err := os.Remove(c.FilePath)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	return err
}

func NewFileCache(filePath string) *fileCache {
	path := strings.TrimSpace(filePath)
	if path == "" {
		path = defautFilePath
	}

	// If the file path is not absolute
	// allow the directory that will contain the cache to be set with environment variables.
	// If KSM_CACHE_DIR is not set, the cache will be created in the current working directory.
	if !filepath.IsAbs(path) {
		if ksmCacheDir := strings.TrimSpace(os.Getenv("KSM_CACHE_DIR")); ksmCacheDir != "" {
			path = filepath.Join(ksmCacheDir, path)
		}
	}

	return &fileCache{FilePath: path}
}

// Memory based cache
type memoryCache struct {
	cache []byte
}

func (c *memoryCache) SaveCachedValue(data []byte) error {
	c.cache = []byte{} // always erase old value
	if len(data) > 0 {
		bytes := make([]byte, len(data))
		copy(bytes, data)
		c.cache = bytes
	}
	return nil
}

func (c *memoryCache) GetCachedValue() ([]byte, error) {
	return c.cache, nil
}

func (c *memoryCache) Purge() error {
	c.cache = []byte{}
	return nil
}

func NewMemoryCache() *memoryCache {
	return &memoryCache{cache: []byte{}}
}
