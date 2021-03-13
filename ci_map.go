package alligotor

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

type ciMap struct {
	m map[string]interface{}
}

func newCiMap() *ciMap {
	return &ciMap{m: make(map[string]interface{})}
}

func (c ciMap) Get(base []string, name string) (b interface{}, ok bool) {
	return c.get(append(base, name))
}

func (c ciMap) get(toIterate []string) (b interface{}, ok bool) {
	// go through map keys and check if key.ToLower() matches, field.ToLower()
	for key := range c.m {
		if !strings.EqualFold(key, toIterate[0]) {
			continue
		}

		val := c.m[key]

		if len(toIterate) == 1 {
			// no separator in the string -> reached end of search string
			return val, true
		}

		// iterate further through nested fields
		valAsMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, false
		}

		return ciMap{m: valAsMap}.get(toIterate[1:])
	}

	return nil, false
}

func (c *ciMap) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&c.m)
}

func (c *ciMap) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &c.m)
}
