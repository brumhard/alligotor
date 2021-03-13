package alligotor

import (
	"os"
	"strings"
)

const (
	envKey              = "env"
	defaultEnvSeparator = "_"
)

// EnvSource is used to read the configuration from environment variables.
// prefix can be defined to look for environment variables with a certain prefix.
// separator is used for nested structs and also for the Prefix.
// As an example:
// If prefix is set to "example", the separator is set to "_" and the config struct's field is named Port,
// it will by default look for the environment variable "EXAMPLE_PORT".
type EnvSource struct {
	prefix    string
	separator string
	envMap    map[string]string
}

// NewEnvSource returns a new EnvSource.
// prefix defines the prefix to be prepended to the automatically generated names when looking for
// the environment variables.
// prefix can be empty.
// It accepts a EnvOption to override the default env separator.
func NewEnvSource(prefix string, opts ...EnvOption) *EnvSource {
	env := &EnvSource{
		prefix:    prefix,
		separator: defaultEnvSeparator,
	}

	for _, opt := range opts {
		opt(env)
	}

	return env
}

// EnvOption takes an EnvSource as input and modifies it.
type EnvOption func(*EnvSource)

// WithEnvSeparator adds a custom separator to an EnvSource struct.
func WithEnvSeparator(separator string) EnvOption {
	return func(env *EnvSource) {
		env.separator = separator
	}
}

// Init initializes the envMap property.
// It should be used right before calling the Read method to load the latest environment variables.
func (s *EnvSource) Init(_ []*Field) error {
	s.envMap = getEnvAsMap()
	return nil
}

// Read reads the saved environment variables from the Init function and returns the set value for a certain field.
// If not value is set in the flags it returns nil.
func (s *EnvSource) Read(field *Field) (interface{}, error) {
	return readEnv(field, s.prefix, s.envMap, s.separator), nil
}

func readEnv(f *Field, prefix string, envMap map[string]string, separator string) []byte {
	if f.Configs[envKey] != "" {
		f.Name = f.Configs[envKey]
	}

	distinctEnvName := f.FullName(separator)
	if prefix != "" {
		distinctEnvName = prefix + separator + distinctEnvName
	}

	envVal, ok := envMap[strings.ToUpper(distinctEnvName)]
	if !ok {
		return nil
	}

	return []byte(envVal)
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
