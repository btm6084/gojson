# gojson

gojson is a collection of tools for extracting meaningful information from a JSON encoded string, including unmarshaling and single-value extraction. gojson is a read only structure. It was born out of a need to extract untrusted JSON into a struct in a type-safe manner.

One of the major problems that caused gojson to be born was a lack of JSON Type homogeneity for a given field. Numbers were often mixed as JSON Numbers and JSON Strings.

The tests for gojson attempt to be illustrative. If you have a question on usage not found here, try reading / modifying the tests to get an idea of how it performs.

JSON Types
============================
JSON has six major types: JSONObject, JSONArray, JSONString, JSONNumber, JSONBoolean, JSONNull

The data found in a JSON field might not always be the type you wish it to be.
Consider the following small example program:
```
package main

import (
	"encoding/json"
	"fmt"

	"github.com/btm6084/gojson"
)

func main() {
	data := [][]byte{
		[]byte(`{"open_new_window": true}`),
		[]byte(`{"open_new_window": 1}`),
		[]byte(`{"open_new_window": "1"}`),
		[]byte(`{"open_new_window": "true"}`),
		[]byte(`{"open_new_window": "t"}`),
		[]byte(`{"open_new_window": 1.0}`),
	}

	type BusinessRules struct {
		OpenNewWindow bool `json:"open_new_window"`
	}

	fmt.Println("Encoding/JSON")
	for _, d := range data {
		var br BusinessRules

		fmt.Println(json.Unmarshal(d, &br), br)
	}

	fmt.Println()
	fmt.Println("gojson")
	for _, d := range data {
		var br BusinessRules

		fmt.Println(gojson.Unmarshal(d, &br), br)
	}
}
```

For each of these we want the same goal: We want BusinessRules.OpenNewWindow to be boolean true. However, if we try to use encoding/json.Unmarshal, all but one of these
will error. When we use gojson.Unmarshal, we get what we actually want: BusinessRules.OpenNewWindow for every case.

Here's the output:
```
Encoding/JSON
<nil> {true}
json: cannot unmarshal number into Go struct field BusinessRules.open_new_window of type bool {false}
json: cannot unmarshal string into Go struct field BusinessRules.open_new_window of type bool {false}
json: cannot unmarshal string into Go struct field BusinessRules.open_new_window of type bool {false}
json: cannot unmarshal string into Go struct field BusinessRules.open_new_window of type bool {false}
json: cannot unmarshal number into Go struct field BusinessRules.open_new_window of type bool {false}

gojson
<nil> {true}
<nil> {true}
<nil> {true}
<nil> {true}
<nil> {true}
<nil> {true}
```

Available Tools
============================

The gojson operations exist to make meaningful JSON data extractions possible in instances where the unmarshaller is insufficient, cumbersome, or simply overkill.

## Unmarshal
gojson offers a custom unmarshaler which is fully compatible with the json.Unmarshaller interface found in encoding/json. The gojson unmarshaler adds some extra capabilities which encoding/json does not.

### Struct Tags

gojson adds support for some new json tags when Unmarshaling

| Flag | Use |
| ---- | --- |
| `required` | An error will be returned if the required key does not exist in the subject JSON
| `nonempty` | An error will be returned if the required key does not exist in the subject JSON OR if it exists, but is the zero value for the json type.

Zero Values are as follows:

| Type | Value |
| ---- | ----- |
| JSONString | ""
| JSONInt | 0
| JSONFloat | 0 or 0.0
| JSONArray | []
| JSONObject | {}
| JSONBool | false
| JSONNull | undefined

GoJSON will also look for the presence of `gojson` tags prior to looking for `json` tags. If the `gojson` tag exists, it will be used for Unmarshalling. Otherwise, the `json` tag will be used. This is useful for times when you wish to have separate behavior when Unmarshaling via gojson, and marshalling via encoding/json.

Example:
```
package main

import (
	"encoding/json"
	"fmt"

	"github.com/btm6084/gojson"
)

type Example struct {
	Product string `json:"-" gojson:"product"`
}

func main() {
	data := []byte(`{"product": "some product identifier"}`)

	var e Example

	gojson.Unmarshal(data, &e)
	fmt.Println(e.Product) // produces `some product identifier`

	m, _ := json.Marshal(e)
	fmt.Println(string(m)) // produces `{}`
}
```

Output:
```
some product identifier
{}
```


### PostUnmarshalJSON

The gojson unmarshaller provides a new interface, PostUnmarshalJSON, defined as follow:
```
PostUnmarshalJSON([]byte, error) error
```

PostUnmarshalJSON is called *after* the unmarshal process has completed, and provides you with the original JSON byte string and any errors that came out of the unmarshal process. The receiver that you defined PostUnmarshalJSON for will be populated for use. This allows you to capture and recover from specific errors, allocate memory for empty slices/maps, react to missing date, perform operations based on the extracted data, or anything else that suits your need.

### UnmarshalStrict
The default Unmarshal process tries to match the data to the container. This means if you have a json string with an integer, and you unmarshal that into an integer field, the conversion will happen for you automatially.

UnmarshalStrict, instead, attempts to match the container to the data, and will return an error if there is a mismatch.

Example:
```
package main

import (
	"fmt"

	"github.com/btm6084/gojson"
)

func main() {
	var container struct{
		Value int
	}

	data := []byte(`{"value": "12345"}`)

	err := gojson.UnmarshalStrict(data, &container)
	fmt.Println(container.Value, err)

	err = gojson.Unmarshal(data, &container)
	fmt.Println(container.Value, err)
}
```

Produces:
```
0 strict standards error, expected int, got string
12345 <nil>
```

## Extract

The Extract* functions are designed to extract simple values from a json byte string without the need to unmarshal the entire structure. Simply pass in the JSON data and the key path, and you will receive the expected data (or an error, if that key does not exist).

* Extract
Extract(JSONData, Key) returns the data at the requested key, or an error if it doesn't exist. The return values are the data (as a byte slice), the JSON type of the data, and and errors.

* ExtractReader
ExtractReader will extract the requested segment and load it into a JSONReader object.

* ExtractString
ExtractString will extract the requested segment and return the value as a string.

* ExtractInt
ExtractString will extract the requested segment and return the value as an int.

* ExtractFloat
ExtractString will extract the requested segment and return the value as a float.

* ExtractBool
ExtractString will extract the requested segment and return the value as a bool.

* ExtractInterface
ExtractString will extract the requested segment and return the value as an interface.

## Interface Type Conversions

| JSON Type | Interface Type |
| --------- | -------------- |
| JSONInt    | int |
| JSONFloat  | float64 |
| JSONString | string |
| JSONBool   | bool |
| JSONNull   | interface{}(nil) |
| JSONArray  | []interface{} |
| JSONObject | map[string]interface{} |

## JSONReader

One of the primary goals of gojson and the supporting tools is to give you a bit more power and influence over the unmarshal process without having to manually handle error conditions for every single field and element. JSONReader is designed to give you power over how you interact with JSON data.

Note that JSONReader parses the entire JSON byte string on instantiation, although subsequent lookups are indexed. This can be slower than you expect / need if you're not doing a large number of extractions / manipulations. If you only need a couple of fields, try Unmarshal or Extract*. If you need to query the object mutiple times, JSONReader might be a good option.

If you know your key is supposed to be an object, use Get.
If you know your key is supposed to be an array, use GetCollection (returns a slice of gojson objects for you to loop over and continue extraction with)
If you know your key is supposed to be an int, use GetInt
If you know your key is supposed to be an array of ints, use GetIntSlice
etc.

The To* functions return the root node's JSON data as the requested type.
* ToBool
* ToBoolSlice
* ToByteSlice
* ToByteSlices
* ToFloat
* ToFloatSlice
* ToInt
* ToInterface
* ToInterfaceSlice
* ToIntSlice
* ToMapStringBool
* ToMapStringBytes
* ToMapStringFloat
* ToMapStringInt
* ToMapStringInterface
* ToMapStringString
* ToString
* ToStringSlice

Get* functions require a key to extract.
* Get
* GetBool
* GetBoolSlice
* GetByteSlice
* GetByteSlices
* GetCollection
* GetFloat
* GetFloatSlice
* GetInt
* GetInterface
* GetInterfaceSlice
* GetIntSlice
* GetMapStringInterface
* GetString
* GetStringSlice



Most data types have a function. Please see jsonreader.go for a full list.

The Get* functions return the requested type for nested values.

As a final note, gojson's Get* functions always return the Zero value if the key doesn't exist. This property, along with gojson's KeyExists() function, allows you to write quick and easy "isEmpty()" functions to check whether the data you received even has the right keys.

Example Program:
```
package main

import (
	"fmt"
	"log"

	"github.com/btm6084/gojson"
)

func main() {
	data := []byte(`{"documents": [{"city_name": "Some City", "postal_codes": ["123.45", 67890, 102.32, "0", true]}]}`)

	type City struct {
		Name     string `json:"city_name"`
		ZipCodes []int  `json:"postal_codes"`
	}

	reader, err := gojson.NewJSONReader(data)
	if err != nil {
		log.Fatal(err)
	}

	city := City{
		Name:     reader.GetString("documents.0.city_name"),
		ZipCodes: reader.GetIntSlice("documents.0.postal_codes"),
	}

	fmt.Println("City Name: ", city.Name)
	fmt.Println("Zip Codes: ", city.ZipCodes)
	fmt.Println()

	fmt.Println("Key                  : Value")
	fmt.Println()
	k := "documents.0.postal_codes"
	fmt.Printf("GetFloatSlice        : [%s]:[%#v] (%T)\n", k, reader.GetFloatSlice(k), reader.GetFloatSlice(k))
	fmt.Printf("GetMapStringInterface: [%s]:[%#v] (%T)\n", k, reader.GetMapStringInterface(k), reader.GetMapStringInterface(k))
	fmt.Println()

	subReader := reader.Get("documents.0.postal_codes")
	for _, k := range subReader.Keys {
		fmt.Printf("GetFloat             : [%s]:[%#v] (%T)\n", k, subReader.GetFloat(k), subReader.GetFloat(k))
		fmt.Printf("GetString            : [%s]:[%#v] (%T)\n", k, subReader.GetString(k), subReader.GetString(k))
		fmt.Printf("GetBool              : [%s]:[%#v] (%T)\n", k, subReader.GetBool(k), subReader.GetBool(k))
		fmt.Printf("GetInt               : [%s]:[%#v] (%T)\n", k, subReader.GetInt(k), subReader.GetInt(k))
		fmt.Println()
	}
	fmt.Println()
}
```

Output:
```
City Name:  Some City
Zip Codes:  [123 67890 102 0 1]

Key                  : Value

GetFloatSlice        : [documents.0.postal_codes]:[[]float64{123.45, 67890, 102.32, 0, 1}] ([]float64)
GetMapStringInterface: [documents.0.postal_codes]:[map[string]interface {}{"0":"123.45", "1":67890, "2":102.32, "3":"0", "4":true}] (map[string]interface {})

GetFloat             : [0]:[123.45] (float64)
GetString            : [0]:["123.45"] (string)
GetBool              : [0]:[false] (bool)
GetInt               : [0]:[123] (int)

GetFloat             : [1]:[67890] (float64)
GetString            : [1]:["67890"] (string)
GetBool              : [1]:[true] (bool)
GetInt               : [1]:[67890] (int)

GetFloat             : [2]:[102.32] (float64)
GetString            : [2]:["102.32"] (string)
GetBool              : [2]:[true] (bool)
GetInt               : [2]:[102] (int)

GetFloat             : [3]:[0] (float64)
GetString            : [3]:["0"] (string)
GetBool              : [3]:[false] (bool)
GetInt               : [3]:[0] (int)

GetFloat             : [4]:[1] (float64)
GetString            : [4]:["true"] (string)
GetBool              : [4]:[true] (bool)
GetInt               : [4]:[1] (int)
```

IsJSON Functions
==============
GoJSON provides a number of Is* functions for use in validating JSON.

* IsJSON
* IsJSONArray
* IsJSONFalse
* IsJSONNull
* IsJSONNumber
* IsJSONObject
* IsJSONString
* IsJSONTrue

Tests
=====

The tests for gojson attempt to be illustrative. If you have a question on usage not found here, try reading / modifying the tests to get an idea of how it performs.

Bonus Example
=============

```
package main

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/btm6084/gojson"
)

func main() {

	rawJSON := `{"value": "17", "name": "My Thing", "children": ["child1", "true", false, 17], "stuff": [ {"value": "4", "name": "stuff1"}, {"value": 5, "name": "stuff2"} ]}`

	type Thing struct {
		Value float64 `json:"value"`
		Name  string  `json:"name"`
	}

	type Things struct {
		Value    float64       `json:"value"`
		Name     string        `json:"name"`
		Children []interface{} `json:"children"`
		Stuff    []Thing       `json:"stuff"`
	}

	reader, err := gojson.NewJSONReader([]byte(rawJSON))
	if err != nil {
		log.Fatal(err)
	}

	t := Things{}

	t.Name = reader.GetString("name")
	t.Value = reader.GetFloat("value")
	t.Children = reader.GetInterfaceSlice("children")

	stuffs := reader.GetCollection("stuff")

	for _, s := range stuffs {
		t.Stuff = append(t.Stuff, Thing{
			s.GetFloat("value"),
			s.GetString("name"),
		})
	}

	fmt.Println(spew.Sdump(t))
}
```

# Output:
```
(main.Things) {
	Value: (float64) 17,
	Name: (string) (len=8) "My Thing",
	Children: ([]interface {}) (len=4 cap=4) {
		(string) (len=6) "child1",
		(string) (len=4) "true",
		(bool) false,
		(int) 17
	},
	Stuff: ([]main.Thing) (len=2 cap=2) {
		(main.Thing) {
			Value: (float64) 4,
			Name: (string) (len=6) "stuff1"
		},
		(main.Thing) {
			Value: (float64) 5,
			Name: (string) (len=6) "stuff2"
		}
	}
}
```