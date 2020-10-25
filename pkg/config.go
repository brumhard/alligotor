package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	tag     = "config"
	envKey  = "env"
	flagKey = "flag"
)

type Collector struct {
	Files ConfigFiles
	Env   bool
	Flags bool
}

type ConfigFiles struct {
	Locations []string
	BaseName  string
}

type ParameterConfig struct {
	// TODO: add support for file option
	EnvName string
	Flag    *Flag
}

type Flag struct {
	Name      string
	ShortName string
}

type Field struct {
	Name   string
	Value  reflect.Value
	Config ParameterConfig
}

func (c *Collector) Get(v interface{}) error {
	// TODO: add recursion
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		// TODO: define package lvl error instead
		return fmt.Errorf("pointer is expected")
	}

	// collect info about fields with tags, value...
	t := reflect.Indirect(value)
	var fields []Field
	for i := 0; i < t.NumField(); i++ {
		configStr, ok := t.Type().Field(i).Tag.Lookup(tag)
		if !ok {
			// TODO: check if field is struct and if so go through fields recursively
			continue
		}

		fieldConfig := ParameterConfig{}

		for _, paramStr := range strings.Split(configStr, ",") {
			keyVal := strings.Split(paramStr, "=")
			if len(keyVal) != 2 {
				panic("invalid config struct tag format")
			}

			key := keyVal[0]
			val := keyVal[1]

			switch key {
			case envKey:
				fieldConfig.EnvName = val
			case flagKey:
				flagConf, err := readFlag(val)
				if err != nil {
					return err
				}

				fieldConfig.Flag = flagConf
			}
		}

		fields = append(fields, Field{
			Name:   t.Type().Field(i).Name,
			Value:  t.Field(i),
			Config: fieldConfig,
		})
	}

	// read files
	for _, fileLocation := range c.Files.Locations {
		fileInfos, err := ioutil.ReadDir(fileLocation)
		if err != nil {
			continue
		}

		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if strings.TrimSuffix(name, path.Ext(name)) != c.Files.BaseName {
				continue
			}

			fileBytes, err := ioutil.ReadFile(path.Join(fileLocation, name))
			if err != nil {
				return err
			}

			m, err := unmarshal(fileBytes)
			if err != nil {
				return err
			}

			for _, f := range fields {
				if valueForField, ok := m.get(f.Name); ok {
					f.Value.Set(reflect.ValueOf(valueForField))
				}
			}
		}
	}

	return nil
}

func unmarshal(bytes []byte) (*ciMap, error) {
	m := newCiMap()
	if err := yaml.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	if err := json.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	return nil, fmt.Errorf("could not unmarshal")
}

func readFlag(flagStr string) (*Flag, error) {
	flagConf := &Flag{}
	flags := strings.Split(flagStr, " ")

	if len(flags) > 2 {
		return nil, fmt.Errorf("malformed flag config strings")
	}

	for _, flag := range flags {
		if len(flag) == 1 {
			flagConf.ShortName = flag
		} else {
			flagConf.Name = flag
		}
	}

	return flagConf, nil
}
