package alligotor

import (
	"encoding"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrPointerExpected    = errors.New("expected a pointer as input")
	ErrUnsupportedType    = errors.New("invalid type")
	ErrCantSet            = errors.New("can't set value")
	ErrDuplicateConfigKey = errors.New("key already used for a config source")
)

const tag = "config"

// DefaultCollector is the default Collector and is used by Get.
var DefaultCollector = &Collector{ // nolint: gochecknoglobals // usage just like in http package
	Sources: []ConfigSource{
		NewFilesSource([]string{"."}, "config"),
		NewEnvSource(""),
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
	Sources []ConfigSource
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
func New(sources ...ConfigSource) *Collector {
	return &Collector{Sources: sources}
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

	for _, source := range c.Sources {
		for _, field := range fields {
			value, err := source.Read(field)
			if err != nil {
				return err
			}

			if value == nil {
				continue
			}

			if err := set(field.value, value); err != nil {
				return err
			}
		}

		if closer, ok := source.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getFieldsConfigsFromValue(value reflect.Value, base ...string) ([]*Field, error) {
	var fields []*Field

	for i := 0; i < value.NumField(); i++ {
		fieldType := value.Type().Field(i)
		fieldValue := reflect.Indirect(value.Field(i))

		fieldConfig, err := readParameterConfig(fieldType.Tag.Get(tag))
		if err != nil {
			return nil, err
		}

		fields = append(fields, &Field{
			Base:    base,
			Name:    fieldType.Name,
			value:   fieldValue,
			Configs: fieldConfig,
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

func readParameterConfig(configStr string) (map[string]string, error) {
	fieldConfig := make(map[string]string)

	if configStr == "" {
		return nil, nil
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

		if _, ok := fieldConfig[key]; ok {
			return nil, errors.Wrap(ErrDuplicateConfigKey, key)
		}

		fieldConfig[key] = val
	}

	return fieldConfig, nil
}

func set(target reflect.Value, value []byte) error {
	receivedev := reflect.New(target.Type())

	var err error
	switch target.Interface().(type) {
	case string:
		value = []byte(strconv.Quote(string(value)))
	case time.Duration:
		dur, err := time.ParseDuration(string(value))
		if err != nil {
			break
		}

		value, err = json.Marshal(dur)
	case []string:
		strSlice := stringSlice{}
		_ = strSlice.UnmarshalText(value)

		value, err = json.Marshal([]string(strSlice))
	case map[string]string:
		strMap := stringMap{}
		_ = strMap.UnmarshalText(value)

		value, err = json.Marshal(map[string]string(strMap))
	}
	if err != nil {
		return err
	}

	if err := json.Unmarshal(value, receivedev.Interface()); err != nil {
		return err
	}

	target.Set(receivedev.Elem())
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
