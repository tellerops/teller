package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/karrick/godirwalk"
	"gopkg.in/yaml.v2"
)

type SnapshotSuite struct {
	Name                 string              `yaml:"name,omitempty"`
	Command              string              `yaml:"command,omitempty"`
	ConfigFileName       string              `yaml:"config_file_name,omitempty"`
	Config               string              `yaml:"config_content,omitempty"`
	InitSnapshot         []SnapshotData      `yaml:"init_snapshot,omitempty"`
	ExpectedSnapshot     []SnapshotData      `yaml:"expected_snapshot,omitempty"`
	ExpectedStdOut       string              `yaml:"expected_stdout,omitempty"`
	ExpectedStdErr       string              `yaml:"expected_stderr,omitempty"`
	ReplaceStdOutContent []ReplaceStdContent `yaml:"replace_stdout_content,omitempty"`
	ReplaceStdErrContent []ReplaceStdContent `yaml:"replace_stderr_content,omitempty"`
}

type SnapshotData struct {
	Path     string `yaml:"path"`
	FileName string `yaml:"file_name"`
	Content  string `yaml:"content"`
}

type ReplaceStdContent struct {
	Search  string `yaml:"search"`
	Replace string `yaml:"replace"`
}

// GetYmlSnapshotSuites returns list of snapshot suite from yml files
func GetYmlSnapshotSuites(folder string) ([]*SnapshotSuite, error) {

	SnapshotSuite := []*SnapshotSuite{}
	err := godirwalk.Walk(folder, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {

			if strings.HasSuffix(osPathname, "yml") {
				snapshotSuite, err := loadSnapshotSuite(osPathname)
				if err != nil {
					return err
				}
				SnapshotSuite = append(SnapshotSuite, snapshotSuite)
			}

			return nil
		},
		Unsorted: true,
	})

	return SnapshotSuite, err
}

// loadSnapshotSuite convert yml file to SnapshotSuite struct
func loadSnapshotSuite(file string) (*SnapshotSuite, error) {
	ymlTestFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	snapshotSuite := &SnapshotSuite{}
	err = yaml.Unmarshal(ymlTestFile, snapshotSuite)
	if err != nil {
		return nil, err
	}
	return snapshotSuite, nil
}

func (s *SnapshotSuite) CrateConfig(dir string) error {
	f, err := os.Create(filepath.Join(dir, s.ConfigFileName))

	if err != nil {
		return err
	}

	defer f.Close()

	t, err := template.New("t").Parse(s.Config)
	if err != nil {
		return err
	}

	type Data struct {
		Folder string
	}

	return t.Execute(f, Data{Folder: dir})
}

// CreateSnapshotData creates filesystem data from the given snapshotData.
// For example, SnapshotData struct descrive the filesystem structure
// └── /folder
//
//	├── settings/
//	│   ├── billing-svc
//	│   └── all/
//	│       ├── foo
//	└── bar
func (s *SnapshotSuite) CreateSnapshotData(snapshotData []SnapshotData, dir string) error {

	for _, data := range snapshotData {
		err := os.MkdirAll(filepath.Join(dir, data.Path), os.ModePerm)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(dir, data.Path, data.FileName), []byte(data.Content), 0644) //nolint
		if err != nil {
			return err
		}
	}
	return nil
}
