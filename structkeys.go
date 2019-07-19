package gojson

import (
	"reflect"
	"strings"
	"sync"
)

// StructKey provides information about a single field in a struct
type StructKey struct {
	// Type is the type of the struct key.
	Type reflect.Type

	// Kind is the kind of the struct key.
	Kind reflect.Kind

	// Name is the primary name for the given struct key.
	Name string

	// Index is the position in the struct where the given field can be found.
	Index int

	// Path maintains the path we need to traverse through struct keys to resolve embeded keys.
	Path []int
}

// StructDescriptor holds parsed metadata about a given struct.
type StructDescriptor struct {
	Keys         map[string]StructKey
	RequiredKeys []string
	NonEmptyKeys []string
}

// NonEmpty returns true if a key is required to be NonEmpty
func (d *StructDescriptor) NonEmpty(key string) bool {
	for _, k := range d.NonEmptyKeys {
		if k == key {
			return true
		}
	}

	return false
}

// Required returns true if a key is required to exist
func (d *StructDescriptor) Required(key string) bool {
	for _, k := range d.RequiredKeys {
		if k == key {
			return true
		}
	}

	return false
}

// Struct Descriptor Cache
// We store the already-processed structs to keep from having to re-process them if
// they come through more than once.
var sdc structDescriptorCache

type structDescriptorCache struct {
	store  map[reflect.Type]*StructDescriptor
	lock   sync.Mutex
	usable bool
}

func (c *structDescriptorCache) Init() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.usable {
		return
	}

	c.store = make(map[reflect.Type]*StructDescriptor)
	c.usable = true
}

func (c *structDescriptorCache) Get(t reflect.Type) *StructDescriptor {
	if !c.usable {
		c.Init()
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if sd, ok := c.store[t]; ok {
		return sd
	}

	return nil
}

func (c *structDescriptorCache) Set(t reflect.Type, sd *StructDescriptor) {
	if !c.usable {
		c.Init()
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.store[t] = sd
}

func getStructInfo(t reflect.Type) *StructDescriptor {
	if c := sdc.Get(t); c != nil {
		return c
	}

	d := &StructDescriptor{}
	d.Keys = make(map[string]StructKey, t.NumField())
	d.RequiredKeys = make([]string, t.NumField())
	d.NonEmptyKeys = make([]string, t.NumField())

	rc := 0
	nc := 0

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Skip non-exported fields.
		if f.PkgPath != "" {
			continue
		}

		// Expand embeded (anonymous) structs.
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			expanded := getStructInfo(f.Type)

			if len(expanded.RequiredKeys) > 0 {
				d.RequiredKeys = append(d.RequiredKeys, expanded.RequiredKeys...)
				rc += len(expanded.RequiredKeys)
			}
			if len(expanded.NonEmptyKeys) > 0 {
				d.NonEmptyKeys = append(d.NonEmptyKeys, expanded.NonEmptyKeys...)
				nc += len(expanded.NonEmptyKeys)
			}

			for n, k := range expanded.Keys {
				k.Path = append([]int{i}, k.Path...)
				d.Keys[n] = k
			}

			continue
		}

		names, required, nonempty := getTags(&f, "json")
		if len(names) == 0 {
			continue
		}

		if required || nonempty {
			d.RequiredKeys[rc] = names[0]
			rc++
		}

		if nonempty {
			d.NonEmptyKeys[nc] = names[0]
			nc++
		}

		d.Keys[names[0]] = StructKey{
			Type:  f.Type,
			Kind:  f.Type.Kind(),
			Name:  names[0],
			Index: i,
		}
	}

	d.RequiredKeys = d.RequiredKeys[:rc]
	d.NonEmptyKeys = d.NonEmptyKeys[:nc]

	sdc.Set(t, d)
	return d
}

// Parse the StructField looking for json tags. If there are no tags, fall back to
// the lowercase of the field name.
func getTags(f *reflect.StructField, key string) ([]string, bool, bool) {
	if len(f.Tag.Get(`json`)) == 0 {
		return []string{strings.ToLower(f.Name)}, false, false
	}

	// We allow gojson tags to be used to separate behavior from encoding/json.
	// A reason to do this would be if you want the field to be unmarshaled, but not
	// marshalled.
	//
	// Example:
	//  type Example struct {
	//      Product *string `json:"-" gojson:"product"`
	//  }
	//
	//  Product would be unmarshalled into the struct as `product`, but json.Marshal would omit it.
	tagSource := "json"
	if f.Tag.Get("gojson") != "" {
		tagSource = "gojson"
	}

	// If the tag consists of ONLY a dash, ignore it.
	if f.Tag.Get(tagSource) == `-` {
		return []string(nil), false, false
	}

	keys := strings.Split(f.Tag.Get(tagSource), `,`)
	final := make([]string, len(keys))

	count := 0
	required := false
	nonempty := false
	for _, k := range keys {
		if strings.ToLower(k) == `omitempty` || k == `` {
			continue
		}

		if strings.ToLower(k) == `required` {
			required = true
			continue
		}

		if strings.ToLower(k) == `nonempty` {
			nonempty = true
			continue
		}

		final[count] = k
		count++
	}

	final = final[:count]
	if len(final) == 0 {
		return []string{strings.ToLower(f.Name)}, false, false
	}

	if len(final) == 1 && final[0] == "-" {
		return []string{}, false, false
	}

	return final, required, nonempty
}
