package alligotor

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
	ErrNoFileFound          = errors.New("no config file could be found")
	ErrUnsupportedType      = errors.New("invalid type")
	ErrCantSet              = errors.New("can't set value")
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

// NewDefault creates a new config collector with default settings.
func NewDefault() *Collector {
	return &Collector{
		Files: FilesConfig{
			Locations: []string{"."},
			BaseName:  "config",
			Separator: defaultFileSeparator,
			Disabled:  false,
		},
		Env: EnvConfig{
			Prefix:    "",
			Separator: defaultEnvSeparator,
			Disabled:  false,
		},
		Flags: FlagsConfig{
			Separator: defaultFlagSeparator,
			Disabled:  false,
		},
	}
}

// Collector is the root struct that implements the main package api.
type Collector struct {
	Files FilesConfig
	Env   EnvConfig
	Flags FlagsConfig
}

// FilesConfig is used to configure the configuration from files.
// Locations can be used to define where to look for files with the defined BaseName.
// Currently only json and yaml files are supported.
// The Separator is used for nested structs.
// If Disabled is true the configuration from files is skipped.
type FilesConfig struct {
	Locations []string
	BaseName  string
	Separator string
	Disabled  bool
}

// EnvConfig is used to configure the configuration from environment variables.
// Prefix can be defined the Collector should look for environment variables with a certain prefix.
// Separator is used for nested structs and also for the Prefix.
// As an example:
// If Prefix is set to "example", the Separator is set to "_" and the config struct's field is named Port,
// the Collector will by default look for the environment variable "EXAMPLE_PORT"
// If Disabled is true the configuration from environment variables is skipped.
type EnvConfig struct {
	Prefix    string
	Separator string
	Disabled  bool
}

// FlagsConfig is used to configure the configuration from command line flags.
// Separator is used for nested structs to construct flag names from parent and child properties recursively.
// If Disabled is true the configuration from flags is skipped.
type FlagsConfig struct {
	Separator string
	Disabled  bool
}

type field struct {
	Base   []string
	Name   string
	Value  reflect.Value
	Config parameterConfig
}

func (f *field) FullName(separator string) string {
	return strings.Join(append(f.Base, f.Name), separator)
}

type parameterConfig struct {
	DefaultFileField string
	DefaultEnvName   string
	Flag             flag
}

type flag struct {
	DefaultName string
	ShortName   string
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

func getFieldsConfigsFromValue(value reflect.Value, base ...string) ([]*field, error) {
	var fields []*field

	for i := 0; i < value.NumField(); i++ {
		fieldType := value.Type().Field(i)
		fieldValue := reflect.Indirect(value.Field(i))

		fieldConfig, err := readParameterConfig(fieldType.Tag.Get(tag))
		if err != nil {
			return nil, err
		}

		fields = append(fields, &field{
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

func readParameterConfig(configStr string) (parameterConfig, error) {
	fieldConfig := parameterConfig{}

	if configStr == "" {
		return parameterConfig{}, nil
	}

	for _, paramStr := range strings.Split(configStr, ",") {
		keyVal := strings.SplitN(paramStr, "=", 2)
		if len(keyVal) != 2 { // nolint: mnd
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
			fieldConfig.DefaultEnvName = val
		case fileKey:
			fieldConfig.DefaultFileField = val
		case flagKey:
			flagConf, err := readFlagConfig(val)
			if err != nil {
				return parameterConfig{}, err
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

func readFiles(fields []*field, config FilesConfig) error {
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

func readFileMap(fields []*field, separator string, m *ciMap) error {
	for _, f := range fields {
		fieldNames := []string{
			f.Config.DefaultFileField,
			f.FullName(separator),
		}

		for _, fieldName := range fieldNames {
			valueForField, ok := m.Get(fieldName)
			if !ok {
				continue
			}

			fieldTypeZero := reflect.Zero(f.Value.Type())
			v := fieldTypeZero.Interface()

			if err := mapstructure.Decode(valueForField, &v); err != nil {
				// if theres a type mismatch check if value is a string and try to use setFromString (e.g. for duration strings)
				if valueString, ok := valueForField.(string); ok {
					if err := setFromString(f.Value, valueString); err != nil {
						return err
					}

					continue
				}

				return err
			}

			f.Value.Set(reflect.ValueOf(v))
		}
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

func readEnv(fields []*field, config EnvConfig, vars map[string]string) error {
	for _, f := range fields {
		destinctEnvName := f.FullName(config.Separator)
		if config.Prefix != "" {
			destinctEnvName = config.Prefix + config.Separator + destinctEnvName
		}

		envNames := []string{
			f.Config.DefaultEnvName,
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

func readPFlags(fields []*field, config FlagsConfig, args []string) error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	fieldToFlagInfo := make(map[*field][]*flagInfo)

	fieldCache := map[string]*flagInfo{}

	for _, f := range fields {
		longName := strings.ToLower(f.FullName(config.Separator))
		defaultName := f.Config.Flag.DefaultName

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

func readFlagConfig(flagStr string) (flag, error) {
	flagConf := flag{}
	flags := strings.Split(flagStr, flagConfigSeparator)

	if len(flags) > 2 { // nolint: gomnd
		return flag{}, ErrMalformedFlagConfig
	}

	for _, f := range flags {
		if len([]rune(f)) == 1 {
			if flagConf.ShortName != "" {
				return flag{}, ErrMalformedFlagConfig
			}

			flagConf.ShortName = f
		} else {
			if flagConf.DefaultName != "" {
				return flag{}, ErrMalformedFlagConfig
			}

			flagConf.DefaultName = f
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
