package yamlpath

import "fmt"

type YamlPathError struct { //nolint:golint
	path    string
	wrapped error
}

func PathError(path string, cause error) *YamlPathError {
	return &YamlPathError{path: path, wrapped: cause}
}

func (e *YamlPathError) Error() string {
	return fmt.Sprintf("Could not traverse path '%s' in Yaml: Cause: %s", e.path, e.wrapped.Error())
}

func (e *YamlPathError) Unwrap() error {
	return e.wrapped
}
