package alligotor

import (
	"os"
	"strings"
)

const (
	envKey              = "env"
	defaultEnvSeparator = "_"
)

// EnvConfig is used to configure the configuration from environment variables.
// Prefix can be defined the Collector should look for environment variables with a certain prefix.
// separator is used for nested structs and also for the Prefix.
// As an example:
// If Prefix is set to "example", the separator is set to "_" and the config struct's field is named Port,
// the Collector will by default look for the environment variable "EXAMPLE_PORT"
// If Disabled is true the configuration from environment variables is skipped.
type EnvSource struct {
	prefix    string
	separator string
	envMap    map[string]string
}

// FromEnvVars is a option for New to enable environment variables as configuration source.
// It takes the prefix for the used environment variables as input parameter.
// FromEnvVars itself takes more options to customize the used separator (WithEnvSeparator).
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

// EnvOption takes an EnvConfig as input and modifies it.
type EnvOption func(*EnvSource)

// WithEnvSeparator adds a custom separator to an EnvConfig struct.
func WithEnvSeparator(separator string) EnvOption {
	return func(env *EnvSource) {
		env.separator = separator
	}
}

func (s *EnvSource) Init(_ []*Field) error {
	s.envMap = getEnvAsMap()
	return nil
}

func (s *EnvSource) Read(field *Field) (interface{}, error) {
	return s.readEnv(field)
}

func (s *EnvSource) readEnv(f *Field) ([]byte, error) {
	distinctEnvName := f.FullName(s.separator)
	if s.prefix != "" {
		distinctEnvName = s.prefix + s.separator + distinctEnvName
	}

	envNames := []string{
		f.Configs[envKey],
		distinctEnvName,
	}

	var finalVal []byte

	for _, envName := range envNames {
		envVal, ok := s.envMap[strings.ToUpper(envName)]
		if !ok {
			continue
		}

		finalVal = []byte(envVal)
	}

	if finalVal == nil {
		return nil, nil
	}

	return finalVal, nil
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
