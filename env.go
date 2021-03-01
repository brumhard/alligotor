package alligotor

import (
	"os"
	"strings"
)

const (
	envKey              = "env"
	defaultEnvSeparator = "_"
)

type Env struct {
	config *EnvConfig
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
}

// FromEnvVars is a option for New to enable environment variables as configuration source.
// It takes the prefix for the used environment variables as input parameter.
// FromEnvVars itself takes more options to customize the used separator (WithEnvSeparator).
func NewEnv(prefix string, opts ...EnvOption) *Env {
	envConfig := &EnvConfig{
		Prefix:    prefix,
		Separator: defaultEnvSeparator,
	}

	for _, opt := range opts {
		opt(envConfig)
	}

	return &Env{config: envConfig}
}

// EnvOption takes an EnvConfig as input and modifies it.
type EnvOption func(*EnvConfig)

// WithEnvSeparator adds a custom separator to an EnvConfig struct.
func WithEnvSeparator(separator string) EnvOption {
	return func(config *EnvConfig) {
		config.Separator = separator
	}
}

func (e *Env) Read(fields []*Field) error {
	return e.readEnv(fields, getEnvAsMap())
}

func (e *Env) readEnv(fields []*Field, vars map[string]string) error {
	for _, f := range fields {
		distinctEnvName := f.FullName(e.config.Separator)
		if e.config.Prefix != "" {
			distinctEnvName = e.config.Prefix + e.config.Separator + distinctEnvName
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
