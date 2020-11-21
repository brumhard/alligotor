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

	"github.com/mitchellh/mapstructure"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	ErrMalformedFlagConfig  = fmt.Errorf("malformed flag config strings")
	ErrFileTypeNotSupported = fmt.Errorf("could not unmarshal file, file type not supported")
	ErrPointerExpected      = fmt.Errorf("expected a pointer as input")
)

const (
	tag     = "config"
	envKey  = "env"
	flagKey = "flag"
	fileKey = "file"

	defaultEnvSeparator  = "_"
	defaultFileSeparator = "."
	defaultFlagSeparator = "-"
)

// TODO: check support for embedded structs
// TODO: clean up linting issues
// TODO: rework to ginkgo tests
// TODO: add support for concurrent tests
// TODO: actually use Disabled flags
type Collector struct {
	Files FilesConfig
	Env   EnvConfig
	Flags FlagsConfig
}

type FilesConfig struct {
	Locations []string
	BaseName  string
	Separator string
	Disabled  bool
}

type EnvConfig struct {
	Prefix    string
	Separator string
	Disabled  bool
}

type FlagsConfig struct {
	Separator string
	Disabled  bool
}

type Field struct {
	Base   []string
	Name   string
	Value  reflect.Value
	Config ParameterConfig
}

func (f *Field) FullName(separator string) string {
	return strings.Join(append(f.Base, f.Name), separator)
}

type ParameterConfig struct {
	FileField string
	EnvName   string
	Flag      Flag
}

type Flag struct {
	Name      string
	ShortName string
}

func (c *Collector) Get(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		return ErrPointerExpected
	}

	// collect info about fields with tags, value...
	t := reflect.Indirect(value)

	fields, err := getFieldsConfigsFromValue(t)
	if err != nil {
		return err
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

func getFieldsConfigsFromValue(value reflect.Value, base ...string) ([]*Field, error) {
	var fields []*Field

	for i := 0; i < value.NumField(); i++ {
		fieldType := value.Type().Field(i)
		fieldValue := value.Field(i)
		configStr := fieldType.Tag.Get(tag)

		fieldConfig := ParameterConfig{}

		if configStr != "" {
			var err error

			fieldConfig, err = readFieldConfig(configStr)
			if err != nil {
				return nil, err
			}
		}

		fields = append(fields, &Field{
			Base:   base,
			Name:   fieldType.Name,
			Value:  fieldValue,
			Config: fieldConfig,
		})

		if fieldValue.Kind() == reflect.Struct {
			newBase := append(base, fieldType.Name)
			subFields, err := getFieldsConfigsFromValue(fieldValue, newBase...)
			if err != nil {
				return nil, err
			}

			fields = append(fields, subFields...)
		}
	}

	return fields, nil
}

func readFieldConfig(configStr string) (ParameterConfig, error) {
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
		case fileKey:
			fieldConfig.FileField = val
		case flagKey:
			flagConf, err := readFlagConfig(val)
			if err != nil {
				return ParameterConfig{}, err
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
			name := fileInfo.Name()
			if strings.TrimSuffix(name, path.Ext(name)) != c.Files.BaseName {
				continue
			}

			fileBytes, err := ioutil.ReadFile(path.Join(fileLocation, name))
			if err != nil {
				return err
			}

			m, err := unmarshal(c.Files.Separator, fileBytes)
			if err != nil {
				return err
			}

			for _, f := range fields {
				fieldName := f.Config.FileField
				if fieldName == "" {
					fieldName = f.FullName(c.Files.Separator)
				}

				valueForField, ok := m.Get(fieldName)
				fieldTypeZero := reflect.Zero(f.Value.Type())
				v := fieldTypeZero.Interface()

				// if value is set, set it, otherwise overwrite with zero value
				// this is to protect values that are set by parent struct but have an overwrite set
				if ok {
					if err := mapstructure.Decode(valueForField, &v); err != nil {
						return err
					}
				}

				f.Value.Set(reflect.ValueOf(v))
			}
		}
	}

	return nil
}

func (c *Collector) readEnv(fields []*Field) error {
	for _, f := range fields {
		envName := f.Config.EnvName
		if envName == "" {
			envName = strings.ToUpper(f.FullName(c.Env.Separator))
		}

		if envVal, ok := os.LookupEnv(envName); ok {
			if err := setFromString(f.Value, envVal); err != nil {
				return err
			}
		}
	}

	return nil
}

type flagInfo struct {
	valueStr *string
	flag     *pflag.Flag
}

func (c *Collector) readPFlags(fields []*Field) error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	fieldToFlagInfo := make(map[*Field]flagInfo)

	for _, f := range fields {
		longName := f.Config.Flag.Name
		if longName == "" {
			longName = strings.ToLower(f.FullName(c.Flags.Separator))
		}

		shortName := ""
		if f.Config.Flag.ShortName != "" {
			shortName = f.Config.Flag.ShortName
		}

		fieldToFlagInfo[f] = flagInfo{
			valueStr: flagSet.StringP(longName, shortName, "", "idk"),
			flag:     flagSet.Lookup(longName),
		}
	}

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return err
	}

	for f, flagInfo := range fieldToFlagInfo {
		// differentiate a flag that is not set from a flag that is set to ""
		if !flagInfo.flag.Changed {
			continue
		}

		if err := setFromString(f.Value, *flagInfo.valueStr); err != nil {
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
		// TODO: add support for bool and other reflect types
	default:
		if textMarshal, ok := target.Interface().(encoding.TextUnmarshaler); ok {
			return textMarshal.UnmarshalText([]byte(value))
		}
		// check if Addr is possible with CanAddr
		if textMarshal, ok := target.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return textMarshal.UnmarshalText([]byte(value))
		}

		target.Set(reflect.ValueOf(value))
	}

	return nil
}

func unmarshal(fileSeperator string, bytes []byte) (*ciMap, error) {
	m := newCiMap(WithSeparator(fileSeperator))
	if err := yaml.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	if err := json.Unmarshal(bytes, m); err == nil {
		return m, nil
	}

	return nil, ErrFileTypeNotSupported
}

func readFlagConfig(flagStr string) (Flag, error) {
	flagConf := Flag{}
	flags := strings.Split(flagStr, " ")

	if len(flags) > 2 {
		return Flag{}, ErrMalformedFlagConfig
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
