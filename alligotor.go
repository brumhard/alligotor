package alligotor

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	ErrPointerExpected    = errors.New("expected a pointer as input")
	ErrStructExpected     = errors.New("expected pointer to struct as input")
	ErrTypeMismatch       = errors.New("type mismatch when trying to assign")
	ErrDuplicateConfigKey = errors.New("key already used for a config source")
)

const (
	configTagKey      = "config"
	descriptionTagKey = "description"
)

// DefaultCollector is the default Collector and is used by Get.
//
//nolint:gochecknoglobals // usage just like in http package
var DefaultCollector = &Collector{
	Sources: []ConfigSource{
		NewFilesSource("./config.*"),
		NewEnvSource(""),
		NewFlagsSource(),
	},
}

// Get is a wrapper around DefaultCollector.Get.
// All predefined configuration sources are enabled.
// For environment variables it uses no prefix and "_" as the separator.
// For flags it use "-" as the separator.
// For config files it uses "config" as the basename and searches in the current directory.
func Get(v interface{}) error {
	return DefaultCollector.Get(v)
}

// Collector is the root struct that implements the main package api.
// The only method that can be called is Collector.Get to unmarshal the found configuration
// values from the configured sources into the provided struct.
// If the default configuration suffices your needs you can just use the package level Get function instead
// without initializing a new Collector struct.
//
// The order in which the different configuration sources overwrite each other can be configured by
// the order in which the sources are defined.
// The default is the following:
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
// duration strings using time.ParseDuration() and time using time.Parse() as well as string slices ([]string)
// in the format val1,val2,val3 and string maps (map[string]string) in the format key1=val1,key2=val2.
type Collector struct {
	Sources []ConfigSource
}

// New returns a new Collector.
// It accepts multiple configuration sources that implement the ConfigSource interface.
// If no sources are present the resulting Collector won't have any configuration sources and return
// the input struct without any changes in the Collector.Get method.
func New(sources ...ConfigSource) *Collector {
	return &Collector{Sources: sources}
}

// Get is the main package function and can be used by its wrapper Get or on a defined Collector struct.
// It expects a pointer to the config struct to write the config variables from the configured source to.
// If the input param is not a pointer to a struct, Get will return an error.
//
// Get looks for config variables in all defined sources.
// Further usage details can be found in the examples or the Collector struct's documentation.
func (c *Collector) Get(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		return ErrPointerExpected
	}

	t := reflect.Indirect(value)
	if t.Kind() != reflect.Struct {
		return ErrStructExpected
	}

	// collect info about fields with tags, value...
	fields, err := getFieldsConfigsFromValue(t, nil)
	if err != nil {
		return err
	}

	for _, source := range c.Sources {
		if initializer, ok := source.(ConfigSourceInitializer); ok {
			if err := initializer.Init(fields); err != nil {
				return err
			}
		}

		for i := range fields {
			fieldVal, err := source.Read(&fields[i])
			if err != nil {
				return err
			}

			if err := set(fields[i].value, fieldVal); err != nil {
				return err
			}
		}
	}

	return nil
}

func getFieldsConfigsFromValue(value reflect.Value, base []Field) ([]Field, error) {
	var fields []Field

	for i := 0; i < value.NumField(); i++ {
		fieldType := value.Type().Field(i)

		fieldValue := reflect.Indirect(value.Field(i))
		if !fieldValue.IsValid() {
			fieldValue = value.Field(i)
		}

		fieldConfig, err := readParameterConfig(fieldType.Tag.Get(configTagKey))
		if err != nil {
			return nil, err
		}

		field := NewField(
			base,
			fieldType.Name,
			fieldType.Tag.Get(descriptionTagKey),
			fieldValue,
			fieldConfig,
		)
		fields = append(fields, field)

		if fieldValue.Kind() == reflect.Struct {
			newBase := append(base, field)

			subFields, err := getFieldsConfigsFromValue(fieldValue, newBase)
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
			return nil, fmt.Errorf("%s: %w", key, ErrDuplicateConfigKey)
		}

		fieldConfig[key] = val
	}

	return fieldConfig, nil
}

func set(target reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	if bytes, ok := value.([]byte); ok {
		if bytes == nil {
			return nil
		}

		var err error

		value, err = fromString(target, string(bytes))
		if err != nil {
			return err
		}
	}

	return trySet(target, reflect.ValueOf(value))
}

func fromString(target reflect.Value, value string) (interface{}, error) {
	specialVal, err := specialTypes(target, value)
	if err != nil {
		return nil, err
	}

	if specialVal != nil {
		return specialVal, nil
	}

	if target.Type().Implements(textUnmarshaler) ||
		(target.CanAddr() && target.Addr().Type().Implements(textUnmarshaler)) {
		// use json capabilities to use TextUnmarshaler interface
		value = strconv.Quote(value)
	}

	receivedev := reflect.New(target.Type())

	if err := json.Unmarshal([]byte(value), receivedev.Interface()); err != nil {
		return nil, err
	}

	return receivedev.Elem().Interface(), nil
}

func specialTypes(target reflect.Value, value string) (finalVal interface{}, err error) {
	switch target.Type() {
	// special cases with special parsing on top of json capabilities
	case durationType:
		return time.ParseDuration(value)
	case durationPtrType:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return nil, err
		}

		return &dur, nil
	case timeType:
		return time.Parse(time.RFC3339, value)
	case timePtrType:
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, err
		}

		return &t, nil
	case stringSliceType:
		strSlice := stringSlice{}
		if err := strSlice.UnmarshalText([]byte(value)); err != nil {
			return nil, err
		}

		return []string(strSlice), nil
	case stringMapType:
		strMap := stringMap{}
		if err := strMap.UnmarshalText([]byte(value)); err != nil {
			return nil, err
		}

		return map[string]string(strMap), nil
	// must not be read by json Unmarshal since that would lead to an error for not quoted string value
	case stringType:
		return value, nil
	case stringPtrType:
		return &value, nil
	}

	return nil, nil
}

func trySet(target, value reflect.Value) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ErrTypeMismatch
		}
	}()

	target.Set(value)

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

type stringSlice []string

func (s *stringSlice) UnmarshalText(text []byte) error {
	tmpSlice := strings.Split(string(text), ",")
	for i := range tmpSlice {
		tmpSlice[i] = strings.TrimSpace(tmpSlice[i])
	}

	*s = tmpSlice

	return nil
}
