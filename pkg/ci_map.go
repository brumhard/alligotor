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

func (c ciMap) set(s string, b bool) {
	c.m[strings.ToLower(s)] = b
}

func (c ciMap) get(s string) (b interface{}, ok bool) {
	b, ok = c.m[strings.ToLower(s)]
	return
}

func (c *ciMap) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]interface{})

	err := value.Decode(&m)
	if err != nil {
		return err
	}

	for key, val := range m {
		c.m[strings.ToLower(key)] = val
	}

	return nil
}

func (c *ciMap) UnmarshalJSON(bytes []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}

	for key, val := range m {
		c.m[strings.ToLower(key)] = val
	}

	return nil
}
