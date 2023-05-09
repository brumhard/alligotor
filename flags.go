package alligotor

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

const (
	flagKey              = "flag"
	flagConfigSeparator  = " "
	defaultFlagSeparator = "."
)

var (
	ErrMalformedFlagConfig = errors.New("malformed flag config strings")
	ErrHelp                = errors.New("help requested")
)

// FlagsSource is used to read the configuration from command line flags.
// separator is used for nested structs to construct flag names from parent and child properties recursively.
type FlagsSource struct {
	separator       string
	fieldToFlagInfo map[string]*flagInfo
}

// NewFlagsSource returns a new FlagsSource.
// It accepts a FlagOption to override the default flag separator.
func NewFlagsSource(opts ...FlagOption) *FlagsSource {
	flags := &FlagsSource{
		separator:       defaultFlagSeparator,
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
		source.separator = separator
	}
}

// Init initializes the fieldToFlagInfos property.
// It should be used right before calling the Read method to load the latest flags.
func (s *FlagsSource) Init(fields []Field) error {
	return s.initFlagMap(fields, os.Args[1:])
}

// Read reads the saved flagSet from the Init function and returns the set value for a certain field.
// If no value is set in the flags it returns nil.
func (s *FlagsSource) Read(field *Field) (interface{}, error) {
	flagInfo, ok := s.fieldToFlagInfo[key(field)]
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

	for i, f := range fields {
		flagConfig, err := readFlagConfig(f.Configs()[flagKey])
		if err != nil {
			return err
		}

		localField := f
		name := extractFlagName(&localField)

		fullname := strings.ToLower(strings.Join(append(f.BaseNames(extractFlagName), name), s.separator))

		s.fieldToFlagInfo[key(&fields[i])] = &flagInfo{
			valueStr: flagSet.StringP(fullname, flagConfig.ShortName, "", f.Description()),
			flag:     flagSet.Lookup(fullname),
		}
	}

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return ErrHelp
		}

		return err
	}

	return nil
}

func extractFlagName(f *Field) string {
	// ignored on this case since the error will be checked in other iterations
	// the fields flagConfigs could be cached to improve performance
	flagConfig, _ := readFlagConfig(f.Configs()[flagKey])
	if flagConfig.LongName != "" {
		return flagConfig.LongName
	}

	return f.Name()
}

type flag struct {
	LongName  string
	ShortName string
}

func key(field *Field) string {
	usualBase := make([]string, 0, len(field.Base()))

	for _, f := range field.Base() {
		usualBase = append(usualBase, f.Name())
	}

	return strings.Join(append(usualBase, field.Name()), "-")
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
