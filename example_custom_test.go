package alligotor_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brumhard/alligotor"
)

type ExampleConfig struct {
	SomeList []string
	SomeMap  map[string]string
	API      struct {
		Enabled bool
	}
	DB struct {
		HostName string
		Timeout  time.Duration
	}
}

func Example_customCollector() {
	// Preconditions
	// Reading only from env vars in this example but it would also be possible
	// to use config file in the current directory (by default).
	// OS args don't have an effect in this example because the source is disabled in the collector.
	os.Args = []string{"cmdName", "--somelist", "a,b,c", "--api-enabled", "true"}
	_ = os.Setenv("EXAMPLE::SOMEMAP", "a=a,b=b,c=c")
	_ = os.Setenv("EXAMPLE::DB::HOSTNAME", "somedb")
	_ = os.Setenv("EXAMPLE::DB::TIMEOUT", "1m0s")

	collector := alligotor.Collector{
		Files: alligotor.FilesConfig{
			Locations: []string{"./", "/etc/example/,", "~/.example/"},
			BaseName:  "example_config",
			Separator: ".",
			Disabled:  false,
		},
		Env: alligotor.EnvConfig{
			Prefix:    "EXAMPLE",
			Separator: "::",
			Disabled:  false,
		},
		Flags: alligotor.FlagsConfig{
			Disabled: true,
		},
	}

	var cfg ExampleConfig
	if err := collector.Get(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg)

	// Output:
	// {[] map[a:a b:b c:c] {false} {somedb 1m0s}}
}
