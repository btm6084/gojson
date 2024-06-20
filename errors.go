package gojson

import "regexp"

var (
	ErrMissingRequiredFields = `%wrequired fields must not be empty [%s]`

	gojsonRequiredKeys = regexp.MustCompile(`(?:nonempty|required) key[s]? '([^']+)'`)
)

func ParseRequiredKeys(err error) string {
	if err == nil {
		return ""
	}
	matches := gojsonRequiredKeys.FindAllStringSubmatch(err.Error(), 1)
	if len(matches) < 1 {
		return ""
	}

	return matches[0][1]
}
