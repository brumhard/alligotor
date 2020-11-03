package pkg

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/pflag"

	"gopkg.in/yaml.v3"
)

const (
	tag     = "config"
	envKey  = "env"
	flagKey = "flag"
	fileKey = "file"
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
	FileField string
	EnvName   string
	Flag      *Flag
}

type Flag struct {
	Name      string
	ShortName string
}

type Field struct {
	Name   string
	Value  reflect.Value
	Config *ParameterConfig
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

	var fields []*Field

	for i := 0; i < t.NumField(); i++ {
		configStr, ok := t.Type().Field(i).Tag.Lookup(tag)
		if !ok {
			// TODO: check if field is struct and if so go through fields recursively
			continue
		}

		fieldConfig, err := readFieldConfig(configStr)
		if err != nil {
			return err
		}

		fields = append(fields, &Field{
			Name:   t.Type().Field(i).Name,
			Value:  t.Field(i),
			Config: fieldConfig,
		})
	}

	// read files
	if err := c.readFiles(fields); err != nil {
		return err
	}

	// read env
	if err := c.readEnv(fields); err != nil {
		return err
	}

	// read flags
	if err := c.readPFlags(fields); err != nil {
		return err
	}

	return nil
}

func readFieldConfig(configStr string) (*ParameterConfig, error) {
	fieldConfig := &ParameterConfig{}

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
		case fileKey:
			fieldConfig.FileField = val
		case flagKey:
			flagConf, err := readFlagConfig(val)
			if err != nil {
				return nil, err
			}

			fieldConfig.Flag = flagConf
		}
	}

	return fieldConfig, nil
}

func (c *Collector) readFiles(fields []*Field) error {
	for _, fileLocation := range c.Files.Locations {
		fileInfos, err := ioutil.ReadDir(fileLocation)
		if err != nil {
			continue
		}

		for _, fileInfo := range fileInfos {
			// TODO: what to do when multiple matching files are found e.g. config.yml & config.json
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
				fieldName := f.Config.FileField
				if fieldName == "" {
					fieldName = f.Name
				}

				if valueForField, ok := m.get(fieldName); ok {
					f.Value.Set(reflect.ValueOf(valueForField))
				}
			}
		}
	}

	return nil
}

func (c *Collector) readEnv(fields []*Field) error {
	// TODO: add support for nested structs
	for _, f := range fields {
		envName := f.Config.EnvName
		if envName == "" {
			envName = strings.ToUpper(f.Name)
		}

		if envVal, ok := os.LookupEnv(envName); ok {
			if err := setFromString(f.Value, envVal); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Collector) readPFlags(fields []*Field) error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	fieldToString := make(map[*Field]*string)
	for _, f := range fields {
		var fieldString string
		flagSet.StringVarP(&fieldString, f.Config.Flag.Name, f.Config.Flag.ShortName, "", "idk")
		fieldToString[f] = &fieldString
	}

	//pflag.Parse()
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return err
	}

	for f, fieldString := range fieldToString {
		if err := setFromString(f.Value, *fieldString); err != nil {
			return err
		}
	}

	return nil
}

var ErrInvalidType = errors.New("invalid type")

func setFromString(target reflect.Value, value string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ErrInvalidType
		}
	}()

	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return err
		}
		target.SetInt(intVal)
	case reflect.String:
		target.SetString(value)
		// TODO: add support for bool and other reflect types?
		// TODO: does the Unmarshaller interface make sense here?, only structs can implement it and in this case
		// the parser would dig into the fields
		// TODO: !!!!!!!!!! could check for struct tags, and if there are some, execute and procedd to fields afterwards
	default:
		if textMarshal, ok := target.Interface().(encoding.TextUnmarshaler); ok {
			if err := textMarshal.UnmarshalText([]byte(value)); err != nil {
				return err
			}
			return nil
		}
		// check if Addr is possible with CanAddr
		if textMarshal, ok := target.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if err := textMarshal.UnmarshalText([]byte(value)); err != nil {
				return err
			}
			return nil
		}

		target.Set(reflect.ValueOf(value))
	}

	return nil
}

func unmarshal(bytes []byte) (*ciMap, error) {
	// TODO: add support for nested structs
	m := newCiMap()
	if err := yaml.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	if err := json.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	return nil, fmt.Errorf("could not unmarshal")
}

func readFlagConfig(flagStr string) (*Flag, error) {
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
