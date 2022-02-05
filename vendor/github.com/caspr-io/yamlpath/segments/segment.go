package segments

import "fmt"

type YamlPathSegment interface {
	NavigateMap(map[string]interface{}) (interface{}, error)
	NavigateArray([]interface{}) (interface{}, error)
}

func NavigateYaml(yaml interface{}, segment YamlPathSegment) (interface{}, error) {
	switch y := yaml.(type) {
	case map[string]interface{}:
		return segment.NavigateMap(y)
	case []interface{}:
		return segment.NavigateArray(y)
	default:
		return nil, fmt.Errorf("no support yet for %v", yaml)
	}
}
