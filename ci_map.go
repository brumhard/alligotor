package alligotor

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultSeparator = "."

type ciMap struct {
	m         map[string]interface{}
	separator string
}

type mapOption func(*ciMap)

func withSeparator(separator string) mapOption {
	return func(c *ciMap) {
		c.separator = separator
	}
}

func newCiMap(options ...mapOption) *ciMap {
	newMap := &ciMap{m: make(map[string]interface{})}

	for _, opt := range options {
		opt(newMap)
	}

	if newMap.separator == "" {
		newMap.separator = defaultSeparator
	}

	return newMap
}

func (c ciMap) Set(s string, b bool) {
	c.m[strings.ToLower(s)] = b
}

func (c ciMap) Get(s string) (b interface{}, ok bool) {
	substr := strings.Split(s, c.separator)

	// go through map keys and check if key.ToLower() matches, field.ToLower()
	for key := range c.m {
		if !strings.EqualFold(key, substr[0]) {
			continue
		}

		val := c.m[key]

		if len(substr) == 1 {
			// no separator in the string -> reached end of search string
			return val, true
		}

		// iterate further through nested fields
		valAsMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, false
		}

		nestedCiMap := ciMap{m: valAsMap, separator: c.separator}

		return nestedCiMap.Get(strings.Join(substr[1:], c.separator))
	}

	return nil, false
}

func (c *ciMap) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&c.m)
}

func (c *ciMap) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &c.m)
}
