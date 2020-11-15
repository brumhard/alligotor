package pkg

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

func (c ciMap) Set(s string, b bool) {
	c.m[strings.ToLower(s)] = b
}

func (c ciMap) Get(s string) (b interface{}, ok bool) {
	// go through map keys and check if key.ToLower() matches, s.ToLower()
	for key := range c.m {
		if !strings.EqualFold(key, s) {
			continue
		}

		return c.m[key], true
	}

	return nil, false

	// TODO: add option for recursive searches
}

func (c *ciMap) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&c.m)
}

func (c *ciMap) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &c.m)
}
