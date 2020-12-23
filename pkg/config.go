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
	"time"

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

	flagConfigSeparator = " "

	defaultEnvSeparator  = "_"
	defaultFileSeparator = "."
	defaultFlagSeparator = "-"
)

// TODO: check support for pointer properties
// TODO: clean up linting issues
// TODO: support duration string in file input
// TODO: if overwrite is set but also normal (for env and file) use normal, what to do for flags?
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

	t := reflect.Indirect(value)

	// collect info about fields with tags, value...
	fields, err := getFieldsConfigsFromValue(t)
	if err != nil {
		return err
	}

	// read files
	if !c.Files.Disabled {
		if err := readFiles(fields, c.Files); err != nil {
			if !errors.Is(err, ErrNoFileFound) {
				fmt.Printf("could not find any files, proceeding with env and flags")

				return err
			}
		}
	}

	// read env
	if !c.Env.Disabled {
		if err := readEnv(fields, c.Env, getEnvAsMap()); err != nil {
			return err
		}
	}

	// read flags
	if !c.Flags.Disabled {
		if err := readPFlags(fields, c.Flags, os.Args[1:]); err != nil {
			return err
		}
	}

	return nil
}

func getFieldsConfigsFromValue(value reflect.Value, base ...string) ([]*Field, error) {
	var fields []*Field

	for i := 0; i < value.NumField(); i++ {
		fieldType := value.Type().Field(i)
		fieldValue := value.Field(i)

		fieldConfig, err := readParameterConfig(fieldType.Tag.Get(tag))
		if err != nil {
			return nil, err
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

func readParameterConfig(configStr string) (ParameterConfig, error) {
	fieldConfig := ParameterConfig{}

	if configStr == "" {
		return ParameterConfig{}, nil
	}

	for _, paramStr := range strings.Split(configStr, ",") {
		keyVal := strings.SplitN(paramStr, "=", 2)
		if len(keyVal) != 2 {
			panic("invalid config struct tag format")
		}

		for _, v := range keyVal {
			if v == "" {
				panic(`config struct tag needs to have the format: config:"file=val,env=val,flag=l long"`)
			}
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
		default:
			panic(
				fmt.Sprintf("only %s, %s, and %s are allowed as config tag keys", envKey, fileKey, flagKey),
			)
		}
	}

	return fieldConfig, nil
}

var ErrNoFileFound = errors.New("no config file could be found")

func readFiles(fields []*Field, config FilesConfig) error {
	fileFound := false

	for _, fileLocation := range config.Locations {
		fileInfos, err := ioutil.ReadDir(fileLocation)
		if err != nil {
			continue
		}

		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if strings.TrimSuffix(name, path.Ext(name)) != config.BaseName {
				continue
			}

			fileFound = true

			fileBytes, err := ioutil.ReadFile(path.Join(fileLocation, name))
			if err != nil {
				return err
			}

			m, err := unmarshal(config.Separator, fileBytes)
			if err != nil {
				return err
			}

			if err := readFileMap(fields, config.Separator, m); err != nil {
				return err
			}
		}
	}

	if !fileFound {
		return ErrNoFileFound
	}

	return nil
}

func readFileMap(fields []*Field, separator string, m *ciMap) error {
	for _, f := range fields {
		fieldName := f.Config.FileField
		if fieldName == "" {
			fieldName = f.FullName(separator)
		}

		valueForField, ok := m.Get(fieldName)
		if !ok {
			continue
		}

		fieldTypeZero := reflect.Zero(f.Value.Type())
		v := fieldTypeZero.Interface()

		if err := mapstructure.Decode(valueForField, &v); err != nil {
			return err
		}

		f.Value.Set(reflect.ValueOf(v))
	}

	return nil
}

func getEnvAsMap() map[string]string {
	envMap := map[string]string{}

	envKeyVal := os.Environ()
	for _, keyVal := range envKeyVal {
		split := strings.SplitN(keyVal, "=", 2)
		envMap[split[0]] = split[1]
	}

	return envMap
}

func readEnv(fields []*Field, config EnvConfig, vars map[string]string) error {
	for _, f := range fields {
		destinctEnvName := f.FullName(config.Separator)
		if config.Prefix != "" {
			destinctEnvName = config.Prefix + config.Separator + destinctEnvName
		}

		envNames := []string{
			f.Config.EnvName,
			destinctEnvName,
		}

		for _, envName := range envNames {
			envVal, ok := vars[strings.ToUpper(envName)]
			if !ok {
				continue
			}

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

func readPFlags(fields []*Field, config FlagsConfig, args []string) error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	fieldToFlagInfo := make(map[*Field][]*flagInfo)

	fieldCache := map[string]*flagInfo{}

	for _, f := range fields {
		longName := strings.ToLower(f.FullName(config.Separator))
		defaultName := f.Config.Flag.Name

		defaultFlag, ok := fieldCache[defaultName]
		if !ok {
			defaultFlag = &flagInfo{
				valueStr: flagSet.StringP(defaultName, "", "", "default"),
				flag:     flagSet.Lookup(defaultName),
			}
			fieldCache[defaultName] = defaultFlag
		}

		fieldToFlagInfo[f] = []*flagInfo{
			defaultFlag,
			{
				valueStr: flagSet.StringP(longName, f.Config.Flag.ShortName, "", "specific"),
				flag:     flagSet.Lookup(longName),
			},
		}
	}

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	for f, flagInfoSlice := range fieldToFlagInfo {
		for _, flagInfo := range flagInfoSlice {
			// differentiate a flag that is not set from a flag that is set to ""
			if !flagInfo.flag.Changed {
				continue
			}

			if err := setFromString(f.Value, *flagInfo.valueStr); err != nil {
				return err
			}
		}
	}

	return nil
}

var (
	ErrUnsupportedType = errors.New("invalid type")
	ErrCantSet         = errors.New("can't set value")
)

func setFromString(target reflect.Value, value string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ErrUnsupportedType
		}
	}()

	if !target.CanSet() {
		return ErrCantSet
	}

	if value == "" {
		zeroValue := reflect.Zero(target.Type())
		target.Set(zeroValue)

		return nil
	}

	var valToSet interface{}

	switch target.Interface().(type) {
	case int, int8, int16, int32, int64:
		intVal, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return err
		}

		target.SetInt(intVal)

		return nil
	case complex64, complex128:
		complexVal, err := strconv.ParseComplex(value, 0)
		if err != nil {
			return err
		}

		target.SetComplex(complexVal)

		return nil
	case uint, uint8, uint16, uint32, uint64:
		uintVal, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return err
		}

		target.SetUint(uintVal)

		return nil
	case float32, float64:
		floatVal, err := strconv.ParseFloat(value, 0)
		if err != nil {
			return err
		}

		target.SetFloat(floatVal)

		return nil
		// TODO: support zap and logrus loglevel
	case time.Duration:
		valToSet, err = time.ParseDuration(value)
	case time.Time:
		valToSet, err = time.Parse(time.RFC3339, value)
	case bool:
		valToSet, err = strconv.ParseBool(value)
	case string:
		valToSet = value
	case []string:
		strSlice := stringSlice{}
		_ = strSlice.UnmarshalText([]byte(value))

		valToSet = []string(strSlice)
	case map[string]string:
		strMap := stringMap{}
		_ = strMap.UnmarshalText([]byte(value))

		valToSet = map[string]string(strMap)
	case encoding.TextUnmarshaler:
		return target.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
	default:
		// check if Addr implements TextUnmarshaler interface
		if t, ok := target.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return t.UnmarshalText([]byte(value))
		}

		valToSet = value
	}

	if err != nil {
		return err
	}

	target.Set(reflect.ValueOf(valToSet))

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
	flags := strings.Split(flagStr, flagConfigSeparator)

	if len(flags) > 2 {
		return Flag{}, ErrMalformedFlagConfig
	}

	for _, flag := range flags {
		if len([]rune(flag)) == 1 {
			if flagConf.ShortName != "" {
				return Flag{}, ErrMalformedFlagConfig
			}

			flagConf.ShortName = flag
		} else {
			if flagConf.Name != "" {
				return Flag{}, ErrMalformedFlagConfig
			}

			flagConf.Name = flag
		}
	}

	return flagConf, nil
}

type stringMap map[string]string

func (m stringMap) UnmarshalText(text []byte) error {
	keyVals := stringSlice{}
	_ = keyVals.UnmarshalText(text)

	for _, keyVal := range keyVals {
		split := strings.SplitN(keyVal, "=", 2)
		m[split[0]] = split[1]
	}

	return nil
}

func (m stringMap) MarshalText() ([]byte, error) {
	keyVals := make([]string, 0, len(m))
	for k, v := range m {
		keyVals = append(keyVals, strings.Join([]string{k, v}, "="))
	}

	return stringSlice(keyVals).MarshalText()
}

type stringSlice []string

func (s *stringSlice) UnmarshalText(text []byte) error {
	*s = append(*s, strings.Split(string(text), ",")...)

	return nil
}

func (s stringSlice) MarshalText() ([]byte, error) {
	return []byte(strings.Join(s, ",")), nil
}
