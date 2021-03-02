package alligotor

import (
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

var ErrMalformedFlagConfig = errors.New("malformed flag config strings")

// FlagsConfig is used to configure the configuration from command line flags.
// separator is used for nested structs to construct flag names from parent and child properties recursively.
// If Disabled is true the configuration from flags is skipped.
type FlagsSource struct {
	separator string
}

func NewFlagsSource(opts ...FlagOption) *FlagsSource {
	flags := &FlagsSource{
		separator: defaultFlagSeparator,
	}

	for _, opt := range opts {
		opt(flags)
	}

	return flags
}

// FlagOption takes a FlagsConfig as input and modifies it.
type FlagOption func(*FlagsSource)

// WithFlagSeparator adds a custom separator to a FlagsConfig struct.
func WithFlagSeparator(separator string) FlagOption {
	return func(flags *FlagsSource) {
		flags.separator = separator
	}
}

func (fl *FlagsSource) Read(fields []*Field) error {
	return fl.readPFlags(fields, os.Args[1:])
}

type flagInfo struct {
	valueStr *string
	flag     *pflag.Flag
}

func (fl *FlagsSource) readPFlags(fields []*Field, args []string) error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}

	fieldToFlagInfo := make(map[*Field][]*flagInfo)
	fieldCache := map[string]*flagInfo{}

	for _, f := range fields {
		flagConfig, err := readFlagConfig(f.Configs[flagKey])
		if err != nil {
			return err
		}

		defaultName := flagConfig.DefaultName
		longName := strings.ToLower(f.FullName(fl.separator))

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
				valueStr: flagSet.StringP(longName, flagConfig.ShortName, "", "specific"),
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

			if err := SetFromString(f.Value(), *flagInfo.valueStr); err != nil {
				return err
			}
		}
	}

	return nil
}

type flag struct {
	DefaultName string
	ShortName   string
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
