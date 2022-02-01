package onepassword

import (
	"encoding/json"
	"errors"
)

type File struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Section     *ItemSection `json:"section,omitempty"`
	Size        int          `json:"size"`
	ContentPath string       `json:"content_path"`
	content     []byte
}

func (f *File) UnmarshalJSON(data []byte) error {
	var jsonFile struct {
		ID          string       `json:"id"`
		Name        string       `json:"name"`
		Section     *ItemSection `json:"section,omitempty"`
		Size        int          `json:"size"`
		ContentPath string       `json:"content_path"`
		Content     []byte       `json:"content,omitempty"`
	}
	if err := json.Unmarshal(data, &jsonFile); err != nil {
		return err
	}
	f.ID = jsonFile.ID
	f.Name = jsonFile.Name
	f.Section = jsonFile.Section
	f.Size = jsonFile.Size
	f.ContentPath = jsonFile.ContentPath
	f.content = jsonFile.Content
	return nil
}

// Content returns the content of the file if they have been loaded and returns an error if they have not been loaded.
// Use `client.GetFileContent(file *File)` instead to make sure the content is fetched automatically if not present.
func (f *File) Content() ([]byte, error) {
	if f.content == nil {
		return nil, errors.New("file content not loaded")
	}
	return f.content, nil
}

func (f *File) SetContent(content []byte) {
	f.content = content
}
