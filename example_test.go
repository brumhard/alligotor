package alligotor_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brumhard/alligotor"
)

type Config struct {
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

func Example() {
	// Preconditions
	// Reading from os args and env vars in this example but it would also be possible
	// to use config file in the current directory (by default).
	os.Args = []string{"cmdName", "--somelist", "a,b,c", "--api-enabled", "true"}
	_ = os.Setenv("SOMEMAP", "a=a,b=b,c=c")
	_ = os.Setenv("DB_HOSTNAME", "somedb")
	_ = os.Setenv("DB_TIMEOUT", "1m0s")

	var cfg Config
	if err := alligotor.Get(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg)

	// Output:
	// {[a b c] map[a:a b:b c:c] {true} {somedb 1m0s}}
}
