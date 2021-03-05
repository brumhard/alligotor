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
// Separator is used for nested structs to construct flag names from parent and child properties recursively.
// If Disabled is true the configuration from flags is skipped.
type FlagsSource struct {
	Separator        string
	fieldToFlagInfos map[string][]*flagInfo
}

func NewFlagsSource(opts ...FlagOption) *FlagsSource {
	flags := &FlagsSource{
		Separator:        defaultFlagSeparator,
		fieldToFlagInfos: make(map[string][]*flagInfo),
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
	return func(source *FlagsSource) {
		source.Separator = separator
	}
}

func (fl *FlagsSource) Init(fields []*Field) error {
	return fl.initFlagMap(fields, os.Args[1:])
}

func (fl *FlagsSource) Read(field *Field) (interface{}, error) {
	flagInfos, ok := fl.fieldToFlagInfos[field.FullName(fl.Separator)]
	if !ok {
		return nil, nil
	}

	var finalVal []byte

	for _, flagInfo := range flagInfos {
		// differentiate a flag that is not set from a flag that is set to ""
		if !flagInfo.flag.Changed {
			continue
		}

		finalVal = []byte(*flagInfo.valueStr)
	}

	if finalVal == nil {
		return nil, nil
	}

	return finalVal, nil
}

type flagInfo struct {
	valueStr *string
	flag     *pflag.Flag
}

func (fl *FlagsSource) initFlagMap(fields []*Field, args []string) error {
	flagSet := pflag.NewFlagSet("config", pflag.ContinueOnError)
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}

	fieldCache := map[string]*flagInfo{}
	for _, f := range fields {
		flagConfig, err := readFlagConfig(f.Configs[flagKey])
		if err != nil {
			return err
		}

		defaultName := flagConfig.DefaultName
		longName := strings.ToLower(f.FullName(fl.Separator))

		defaultFlag, ok := fieldCache[defaultName]
		if !ok {
			defaultFlag = &flagInfo{
				valueStr: flagSet.StringP(defaultName, "", "", "default"),
				flag:     flagSet.Lookup(defaultName),
			}
			fieldCache[defaultName] = defaultFlag
		}

		fl.fieldToFlagInfos[f.FullName(fl.Separator)] = []*flagInfo{
			defaultFlag,
			{
				valueStr: flagSet.StringP(longName, flagConfig.ShortName, "", "specific"),
				flag:     flagSet.Lookup(longName),
			},
		}
	}

	return flagSet.Parse(args)
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
