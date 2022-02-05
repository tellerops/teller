package yamlpath

import (
	"strings"

	"github.com/caspr-io/yamlpath/segments"
)

// https://pypi.org/project/yamlpath/#supported-yaml-path-segments

// YamlPath traverses the yaml document to and returns the retrieved value
func YamlPath(yaml map[string]interface{}, path string) (interface{}, error) {
	splitPath, err := segments.ParseSegment(path)
	if err != nil {
		return nil, PathError(path, err)
	}
	// fmt.Printf("%v, %d", splitPath, len(splitPath))

	var value interface{} = yaml

	for _, pathPart := range splitPath {
		returned, err := segments.NavigateYaml(value, pathPart)
		if err != nil {
			return nil, PathError(path, err)
		}

		value = returned
	}

	return value, nil
}

// func navigateYaml(yaml interface{}, part string) (interface{}, error) {
// 	switch y := yaml.(type) {
// 	case map[string]interface{}:
// 		return navigateMap(y, part)
// 	case []interface{}:
// 		return navigateArray(y, part)
// 	default:
// 		return nil, fmt.Errorf("no support yet for %v", yaml)
// 	}
// }

// func navigateArray(l []interface{}, part string) (interface{}, error) {
// switch {
// case regexps[ExplicitIndexPart].MatchString(part):
// 	i, err := strconv.Atoi(part[1 : len(part)-1])
// 	if err != nil {
// 		return nil, err
// 	}

// 	if i < len(l) {
// 		return l[i], nil
// 	}

// 	return nil, fmt.Errorf("out of bounds '%d' for array of length '%d'", i, len(l))
// case regexps[SlicePart].MatchString(part):
// case regexps[ImplicitIndexPart].MatchString(part):
// 	i, err := strconv.Atoi(part)
// 	if err != nil {
// 		return nil, fmt.Errorf("part '%s' is not an index into an array. %w", part, err)
// 	}

// 	return l[i], nil
// case regexps[KeyPart].MatchString(part):
// case regexps[ValueSearchPart].MatchString(part):
// 	toFind := part[3 : len(part)-1]
// 	operator := part[2]

// 	for _, i := range l {
// 		switch s := i.(type) {
// 		case string:
// 			if valueMatches(s, toFind, operator) {
// 				return s, nil
// 			}

// 			continue
// 		default:
// 			return nil, fmt.Errorf("could not search for value '%s' as list does not contain strings", part)
// 		}
// 	}

// 	return nil, fmt.Errorf("could not find match for search part '%s'", part)
// default:
// 	return nil, fmt.Errorf("part '%s' not supported for array", part)
// }
// }

func valueMatches(s string, find string, operator byte) bool {
	switch operator {
	case '^':
		return strings.HasPrefix(s, find)
	case '$':
		return strings.HasSuffix(s, find)
	case '%':
		return strings.Contains(s, find)
	default:
		return false
	}
}

// func navigateMap(m map[string]interface{}, part string) (interface{}, error) {
// 	switch {
// 	case regexps[KeySearchPart].MatchString(part):
// 		key := part[3 : len(part)-1]
// 		return m[key], nil
// 	default:
// 		return nil, fmt.Errorf("no support for part '%s'", part)
// 	}
// }
