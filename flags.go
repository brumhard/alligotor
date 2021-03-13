package alligotor

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	flagKey              = "flag"
	flagConfigSeparator  = " "
	defaultFlagSeparator = "-"
)

var (
	ErrMalformedFlagConfig = errors.New("malformed flag config strings")
	ErrNotFound            = errors.New("not found")
)

// FlagsSource is used to read the configuration from command line flags.
// Separator is used for nested structs to construct flag names from parent and child properties recursively.
type FlagsSource struct {
	Separator       string
	fieldToFlagInfo map[string]*flagInfo
}

// NewFlagsSource returns a new FlagsSource.
// It accepts a FlagOption to override the default flag separator.
func NewFlagsSource(opts ...FlagOption) *FlagsSource {
	flags := &FlagsSource{
		Separator:       defaultFlagSeparator,
		fieldToFlagInfo: make(map[string]*flagInfo),
	}

	for _, opt := range opts {
		opt(flags)
	}

	return flags
}

// FlagOption takes a FlagsSource as input and modifies it.
type FlagOption func(*FlagsSource)

// WithFlagSeparator adds a custom separator to a FlagsSource struct.
func WithFlagSeparator(separator string) FlagOption {
	return func(source *FlagsSource) {
		source.Separator = separator
	}
}

// Init initializes the fieldToFlagInfos property.
// It should be used right before calling the Read method to load the latest flags.
func (s *FlagsSource) Init(fields []Field) error {
	return s.initFlagMap(fields, os.Args[1:])
}

// Read reads the saved flagSet from the Init function and returns the set value for a certain field.
// If not value is set in the flags it returns nil.
func (s *FlagsSource) Read(field Field) (interface{}, error) {
	flagInfo, ok := s.fieldToFlagInfo[field.FullName(s.Separator)]
	if !ok {
		return nil, nil
	}

	if !flagInfo.flag.Changed {
		return nil, nil
	}

	return []byte(*flagInfo.valueStr), nil
}

type flagInfo struct {
	valueStr *string
	flag     *pflag.Flag
}

func (s *FlagsSource) initFlagMap(fields []Field, args []string) error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: false}
	flagSet.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])

		flagSet.VisitAll(func(f *pflag.Flag) {
			line := ""
			if f.Shorthand != "" {
				line = fmt.Sprintf("  -%s, --%s", f.Shorthand, f.Name)
			} else {
				line = fmt.Sprintf("      --%s", f.Name)
			}

			if f.Usage != "" {
				line = line + ": " + f.Usage
			}

			_, _ = fmt.Fprint(os.Stderr, line)
			_, _ = fmt.Fprintf(os.Stderr, "\n")
		})
	}

	for _, f := range fields {
		mapKey := f.FullName(s.Separator)

		flagConfig, err := readFlagConfig(f.configs[flagKey])
		if err != nil {
			return err
		}

		if flagConfig.LongName != "" {
			f.name = flagConfig.LongName
		}

		longName := strings.ToLower(f.FullName(s.Separator))

		s.fieldToFlagInfo[mapKey] = &flagInfo{
			valueStr: flagSet.StringP(longName, flagConfig.ShortName, "", f.description),
			flag:     flagSet.Lookup(longName),
		}
	}

	if err := flagSet.Parse(args); err != nil {
		if !errors.Is(err, pflag.ErrHelp) {
			return err
		}
	}

	return nil
}

type flag struct {
	LongName  string
	ShortName string
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
			if flagConf.LongName != "" {
				return flag{}, ErrMalformedFlagConfig
			}

			flagConf.LongName = f
		}
	}

	return flagConf, nil
}
