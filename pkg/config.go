package pkg

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	tag        = "config"
	defaultKey = "default"
	envKey     = "env"
	flagKey    = "flag"
)

type Collector struct {
	Files ConfigFiles
	Env   bool
}

type ConfigFiles struct {
	Locations []string
	BaseName  string
}

type Parameter struct {
	DefaultStr string
	EnvName    string
	Flag       *Flag
}

type Flag struct {
	Name      string
	ShortName string
}

func (c *Collector) Get(v interface{}) error {
	// TODO: read from file
	// TODO: check that v is pointer
	// TODO: dereference v below
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		// TODO: define package lvl error instead
		return fmt.Errorf("pointer is expected")
	}

	t := reflect.Indirect(value).Type()
	configMap := make(map[*reflect.StructField]Parameter)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		configStr, ok := field.Tag.Lookup(tag)
		if !ok {
			// TODO: check if field is struct and if so go through fields recursively
			continue
		}

		fieldConfig := Parameter{}

		for _, paramStr := range strings.Split(configStr, ",") {
			keyVal := strings.Split(paramStr, "=")
			if len(keyVal) != 2 {
				panic("invalid config struct tag format")
			}

			key := keyVal[0]
			val := keyVal[1]

			switch key {
			case defaultKey:
				fieldConfig.DefaultStr = val
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

		configMap[&field] = fieldConfig
	}

	return nil
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
