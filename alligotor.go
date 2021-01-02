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
	ErrMalformedFlagConfig  = errors.New("malformed flag config strings")
	ErrFileTypeNotSupported = errors.New("could not unmarshal file, file type not supported")
	ErrPointerExpected      = errors.New("expected a pointer as input")
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

// DefaultCollector is the default Collector and is used by Get.
var DefaultCollector = &Collector{ // nolint: gochecknoglobals // usage just like in http package
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

// Get is a wrapper around DefaultCollector.Get.
// All configuration sources are enabled.
// For environment variables it uses no prefix and "_" as the separator.
// For flags it use "-" as the separator.
// For config files it uses "config" as the basename and searches in the current directory.
// It uses "." as the separator.
func Get(v interface{}) error {
	return DefaultCollector.Get(v)
}

// Collector is the root struct that implements the main package api.
// The only method that can be called is Collector.Get to unmarshal the found configuration
// values from the configured sources into the provided struct.
// If the default configuration suffices your needs you can just use the package level Get function instead
// without initializing a new Collector struct.
//
// The order in which the different configuration sources overwrite each other is the following:
// defaults -> config files -> environment variables -> command line flags
// (each source is overwritten by the following source)
//
// To define defaults for the config variables it can just be predefined in the struct that the
// configuration is supposed to be unmarshalled into. Properties that are not set in any of
// the configuration sources will keep the preset value.
//
// Since environment variables and flags are purely text based it also supports types that implement
// the encoding.TextUnmarshaler interface like for example zapcore.Level and logrus.Level.
// On top of that custom implementations are already baked into the package to support
// duration strings using time.ParseDuration() as well as string slices ([]string) in the format val1,val2,val3
// and string maps (map[string]string) in the format key1=val1,key2=val2.
type Collector struct {
	Files FilesConfig
	Env   EnvConfig
	Flags FlagsConfig
}

// New returns a new Collector.
// Various options can be used to customize the result.
// If no options are present the resulting Collector won't have any configuration sources and return
// the input struct without any changes in the Collector.Get method.
// Available options are:
// - FromFiles to configure configuration files as input source
// - FromEnvVars to configure environment variables as input source
// - FromCLIFlags to configure command line flags as input source
// Each of these options has an option itself to provide a custom separator.
// They are named WithFileSeparator, WithEnvSeparator and WithFlagSeparator.
func New(opts ...CollectorOption) *Collector {
	c := &Collector{
		Files: FilesConfig{
			Disabled: true,
		},
		Env: EnvConfig{
			Disabled: true,
		},
		Flags: FlagsConfig{
			Disabled: true,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CollectorOption takes a Collector as input and modifies it.
type CollectorOption func(*Collector)

// FromFiles is a option for New to enable configuration files as configuration source.
// It takes the locations/ dirs where to look for files and the basename (without file extension) as input parameters.
// FromFiles itself takes more options to customize the used separator (WithFileSeparator).
func FromFiles(locations []string, baseName string, opts ...FileOption) CollectorOption {
	return func(collector *Collector) {
		collector.Files = FilesConfig{
			Locations: locations,
			BaseName:  baseName,
			Separator: defaultFileSeparator,
			Disabled:  false,
		}

		for _, opt := range opts {
			opt(&collector.Files)
		}
	}
}

// FileOption takes a FilesConfig as input and modifies it.
type FileOption func(*FilesConfig)

// WithFileSeparator adds a custom separator to a FilesConfig struct.
func WithFileSeparator(separator string) FileOption {
	return func(config *FilesConfig) {
		config.Separator = separator
	}
}

// FromEnvVars is a option for New to enable environment variables as configuration source.
// It takes the prefix for the used environment variables as input parameter.
// FromEnvVars itself takes more options to customize the used separator (WithEnvSeparator).
func FromEnvVars(prefix string, opts ...EnvOption) CollectorOption {
	return func(collector *Collector) {
		collector.Env = EnvConfig{
			Prefix:    prefix,
			Separator: defaultEnvSeparator,
			Disabled:  false,
		}

		for _, opt := range opts {
			opt(&collector.Env)
		}
	}
}

// EnvOption takes an EnvConfig as input and modifies it.
type EnvOption func(*EnvConfig)

// WithEnvSeparator adds a custom separator to an EnvConfig struct.
func WithEnvSeparator(separator string) EnvOption {
	return func(config *EnvConfig) {
		config.Separator = separator
	}
}

// FromCLIFlags is a option for New to enable command line flags as configuration source.
// FromCLIFlags itself takes more options to customize the used separator (WithFlagSeparator).
func FromCLIFlags(opts ...FlagOption) CollectorOption {
	return func(collector *Collector) {
		collector.Flags = FlagsConfig{
			Separator: defaultFlagSeparator,
			Disabled:  false,
		}

		for _, opt := range opts {
			opt(&collector.Flags)
		}
	}
}

// FlagOption takes a FlagsConfig as input and modifies it.
type FlagOption func(*FlagsConfig)

// WithFlagSeparator adds a custom separator to a FlagsConfig struct.
func WithFlagSeparator(separator string) FlagOption {
	return func(config *FlagsConfig) {
		config.Separator = separator
	}
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

// Get is the main package function and can be used by its wrapper Get or on a defined Collector struct.
// It expects a pointer to the config struct to write the config variables from the configured source to.
// If the input param is not a pointer, Get will return an error.
//
// Get looks for config variables all sources that are not disabled.
// Further usage details can be found in the examples or the Collector struct's documentation.
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

				// if the target is a struct there are also fields for the child properties and it should be tried
				// to set these before returning an error
				if f.Value.Kind() == reflect.Struct {
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
		distinctEnvName := f.FullName(config.Separator)
		if config.Prefix != "" {
			distinctEnvName = config.Prefix + config.Separator + distinctEnvName
		}

		envNames := []string{
			f.Config.DefaultEnvName,
			distinctEnvName,
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
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}

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

func setFromString(target reflect.Value, value string) (err error) { // nolint: funlen,gocyclo // just huge switch case
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

func unmarshal(fileSeparator string, bytes []byte) (*ciMap, error) {
	m := newCiMap(withSeparator(fileSeparator))
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

	if len(flags) > 2 {
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
		for i := range split {
			split[i] = strings.TrimSpace(split[i])
		}

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
	tmpSlice := strings.Split(string(text), ",")
	for i := range tmpSlice {
		tmpSlice[i] = strings.TrimSpace(tmpSlice[i])
	}

	*s = tmpSlice

	return nil
}

func (s stringSlice) MarshalText() ([]byte, error) {
	return []byte(strings.Join(s, ",")), nil
}
