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

func (e *EnvSource) Read(fields []*Field) error {
	return e.readEnv(fields, getEnvAsMap())
}

func (e *EnvSource) readEnv(fields []*Field, vars map[string]string) error {
	for _, f := range fields {
		distinctEnvName := f.FullName(e.separator)
		if e.prefix != "" {
			distinctEnvName = e.prefix + e.separator + distinctEnvName
		}

		envNames := []string{
			f.Configs[envKey],
			distinctEnvName,
		}

		for _, envName := range envNames {
			envVal, ok := vars[strings.ToUpper(envName)]
			if !ok {
				continue
			}

			if err := SetFromString(f.Value(), envVal); err != nil {
				return err
			}
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
