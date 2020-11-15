package pkg

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

const separator = "."

type ciMap struct {
	m map[string]interface{}
}

func newCiMap() *ciMap {
	return &ciMap{m: make(map[string]interface{})}
}

func (c ciMap) Set(s string, b bool) {
	c.m[strings.ToLower(s)] = b
}

func (c ciMap) Get(s string) (b interface{}, ok bool) {
	substr := strings.Split(s, separator)
	field := substr[0]

	// go through map keys and check if key.ToLower() matches, field.ToLower()
	for key := range c.m {
		if !strings.EqualFold(key, field) {
			continue
		}

		val := c.m[field]

		if len(substr) == 1 {
			// no separator in the string -> reached end of search string
			return val, true
		}

		// iterate further through nested fields
		valAsMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, false
		}

		nestedCiMap := ciMap{m: valAsMap}

		return nestedCiMap.Get(strings.Join(substr[1:], separator))
	}

	return nil, false
}

func (c *ciMap) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&c.m)
}

func (c *ciMap) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &c.m)
}
