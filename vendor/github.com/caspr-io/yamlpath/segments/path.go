package segments

import "fmt"

func ParseSegment(path string) ([]YamlPathSegment, error) {
	if path[0] == '/' {
		return parseSegment(path[1:], '/')
	}

	return parseSegment(path, '.')
}

func parseSegment(path string, separator rune) ([]YamlPathSegment, error) {
	segments := []YamlPathSegment{}
	currentSegment := []rune{}
	i := 0

	for i < len(path) {
		r := rune(path[i])
		switch r {
		case separator:
			if err := addSegment(currentSegment, &segments); err != nil {
				return nil, err
			}

			currentSegment = []rune{}
		case '\\':
			i++
			r = rune(path[i])
			currentSegment = append(currentSegment, r)
		case '"', '\'':
			if err := addSegment(currentSegment, &segments); err != nil {
				return nil, err
			}

			currentSegment = []rune{}

			p, endIdx, err := parsePathUntil(path, i+1, r, false) //nolint:gomnd
			if err != nil {
				return nil, err
			}

			segments = append(segments, p)
			i = endIdx
		case '[':
			if err := addSegment(currentSegment, &segments); err != nil {
				return nil, err
			}

			currentSegment = []rune{}

			p, endIdx, err := parsePathUntil(path, i, ']', true)
			if err != nil {
				return nil, err
			}

			segments = append(segments, p)
			i = endIdx
		default:
			currentSegment = append(currentSegment, r)
		}
		i++
	}

	if len(currentSegment) > 0 {
		if err := addSegment(currentSegment, &segments); err != nil {
			return nil, err
		}
	}

	return segments, nil
}

func addSegment(segment []rune, segments *[]YamlPathSegment) error {
	if len(segment) == 0 {
		return nil
	}

	p, err := DetectSegment(string(segment))
	if err != nil {
		return err
	}

	l := append(*segments, p)
	*segments = l

	return nil
}

func parsePathUntil(path string, idx int, stopOn rune, inclusive bool) (YamlPathSegment, int, error) {
	segment := []rune{}
	i := idx

	for i < len(path) {
		r := rune(path[i])
		segment = append(segment, r)

		if r == stopOn {
			if !inclusive {
				segment = segment[0 : len(segment)-1]
			}

			ypp, err := DetectSegment(string(segment))

			return ypp, i + 1, err
		}
		i++
	}

	return nil, -1, fmt.Errorf("could not find terminating '%c' in path '%s'", stopOn, path[idx:])
}
