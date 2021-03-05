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
		Enabled *bool
	}
	DB struct {
		HostName *string
		Timeout  *time.Duration
	}
	TimeStamp *time.Time
}

func Example() {
	// Preconditions
	// Reading from os args and env vars in this example but it would also be possible
	// to use config file in the current directory (by default).
	os.Args = []string{"cmdName", "--somelist", "a,b,c", "--api-enabled", "true"}
	_ = os.Setenv("SOMEMAP", "a=a,b=b,c=c")
	_ = os.Setenv("DB_HOSTNAME", "somedb")
	_ = os.Setenv("DB_TIMEOUT", "1m0s")
	_ = os.Setenv("TIMESTAMP", "2007-01-02T15:04:05Z")

	var cfg Config
	if err := alligotor.Get(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.SomeList, cfg.SomeMap, *cfg.API.Enabled, *cfg.DB.HostName, *cfg.DB.Timeout, cfg.TimeStamp.UTC())

	// Output:
	// [a b c] map[a:a b:b c:c] true somedb 1m0s 2007-01-02 15:04:05 +0000 UTC
}
