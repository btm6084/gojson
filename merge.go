package gojson

import (
	"bytes"
	"errors"
	"strings"
)

var (
	// ErrMergeType is returned when MergeJSON receives anything other than JSONObject or JSONArray
	ErrMergeType = errors.New("only JSONObject can be merged")
)

// MergeJSON merges two JSONObject together and returns the combination.
// Base and Patch must be type JSONObject.
// When keys exist in the base but not in the patch, it will be retained.
// When keys exist in the patch but not in the base, it will be added to the base.
// JSONArray, JSONString, JSONInt, JSONFloat, JSONNull in the patch will always override the base.
// JSONObject in the patch will override anything but JSONObject in the base.
// JSONObject in the patch AND JSONObject in the base will be merged together.
func MergeJSON(base, patch []byte) ([]byte, error) {
	a, err := NewJSONReader(base)
	if err != nil {
		return nil, err
	}

	b, err := NewJSONReader(patch)
	if err != nil {
		return nil, err
	}

	if a.Type != JSONObject || b.Type != JSONObject {
		return nil, ErrMergeType
	}

	p := merge(a.parsed, b.parsed)
	return toByteString(p, a.Type, uniqueString(append(a.Keys, b.Keys...), false)), nil
}

func merge(a, b map[string]parsed) map[string]parsed {
	// If B is empty, there's nothing to merge.
	if len(b) == 0 {
		return a
	}

	if len(a) == 0 {
		return b
	}

	i, d := id(keys(a), keys(b))

	// For each key in b not in a, add it to a
	for _, k := range d {
		a[k] = b[k]
	}

	// For each key that matches between a and b, merge them.
	for _, k := range i {

		if a[k].dtype == JSONObject && b[k].dtype == JSONObject {
			p := a[k]
			p.children = merge(a[k].children, b[k].children)
			p.keys = uniqueString(append(p.keys, b[k].keys...), false)
			a[k] = p
			continue
		}

		a[k] = b[k]
	}

	return a
}

func toByteString(p map[string]parsed, t string, keys []string) []byte {
	if len(p) == 0 {
		return nil
	}

	contents := make([]string, len(keys))

	open, close := `[`, `]`
	if t == JSONObject {
		open, close = `{`, `}`
	}

	i := 0
	for _, k := range keys {
		v := p[k]
		buf := bytes.NewBuffer([]byte{})

		switch t {
		case JSONObject:
			switch v.dtype {
			case JSONObject, JSONArray:
				if IsEmptyObject(v.bytes) {
					buf.WriteString(`"` + k + `":{}`)
					contents[i] = buf.String()
					break
				}
				if IsEmptyArray(v.bytes) {
					buf.WriteString(`"` + k + `":[]`)
					contents[i] = buf.String()
					break
				}

				b := toByteString(v.children, v.dtype, v.keys)
				if b == nil {
					// Skip this key.
					continue
				}

				buf.WriteString(`"` + k + `":` + string(b))
				contents[i] = buf.String()
			case JSONString:
				buf.WriteString(`"` + k + `":"` + string(v.bytes) + `"`)
				contents[i] = buf.String()
			case JSONInvalid:
			default:
				buf.WriteString(`"` + k + `":` + string(v.bytes))
				contents[i] = buf.String()
			}
		case JSONArray:
			switch v.dtype {
			case JSONObject, JSONArray:
				if IsEmptyObject(v.bytes) {
					buf.WriteString(`{}`)
					contents[i] = buf.String()
					break
				}
				if IsEmptyArray(v.bytes) {
					buf.WriteString(`[]`)
					contents[i] = buf.String()
					break
				}

				b := toByteString(v.children, v.dtype, v.keys)
				if b == nil {
					// Skip this key.
					continue
				}

				buf.WriteString(string(b))
				contents[i] = buf.String()
			case JSONString:
				buf.WriteString(`"` + string(v.bytes) + `"`)
				contents[i] = buf.String()
			case JSONInvalid:
			default:
				buf.WriteString(string(v.bytes))
				contents[i] = buf.String()
			}
		}

		i++
	}

	return []byte(open + strings.Join(contents, ",") + close)
}

func keys(in map[string]parsed) []string {
	out := make([]string, len(in))

	i := 0
	for k := range in {
		out[i] = k
		i++
	}

	return out
}

// id finds the common and not common elements between two slices of string.
// Intersect is common between both.
// Difference exists in b but not in a.
func id(a, b []string) (intersect []string, difference []string) {
	seen := make(map[string]bool)
	for _, v := range a {
		seen[v] = true
	}

	for _, v := range b {
		if _, ok := seen[v]; ok {
			intersect = append(intersect, v)
			continue
		}

		difference = append(difference, v)
	}

	return intersect, difference
}

func uniqueString(in []string, allowEmpty bool) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range in {
		if _, isset := seen[v]; isset {
			continue
		}

		if !allowEmpty && v == "" {
			continue
		}

		seen[v] = true
		out = append(out, v)
	}

	return out
}
