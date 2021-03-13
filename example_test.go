package alligotor_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/brumhard/alligotor"
)

type Config struct {
	SomeList       []string          `description:"list that can ba assigned in format: a,b,c"` // description used for flag usage
	SomeMap        map[string]string `config:"env=MAP"`                                         // overwrite the env var to read
	SomeCustomType SomeCustomType    `config:"env=custom"`                                      // implements text.Unmarshaler
	API            struct {
		Enabled *bool
	}
	DB struct {
		HostName string        `config:"flag=h host"` // overwrite the cli flags to read, h is shortname
		Timeout  time.Duration `config:"flag=time"`   // only overwrite long name
	}
	TimeStamp time.Time `config:"file=custom"` // implements text.Unmarshaler, overwrite key in file
}

func Example() {
	os.Args = []string{"cmdName", "--somelist", "a,b,c", "--api-enabled", "true", "-h", "dbhost"}
	_ = os.Setenv("MAP", "a=a,b=b,c=c")
	_ = os.Setenv("DB_TIMEOUT", "1m0s")
	_ = os.Setenv("TIMESTAMP", "2007-01-02T15:04:05Z")
	_ = os.Setenv("CUSTOM", "key=value")

	var cfg Config
	if err := alligotor.Get(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.SomeList, cfg.SomeMap, cfg.SomeCustomType, *cfg.API.Enabled, cfg.DB.HostName, cfg.DB.Timeout, cfg.TimeStamp.UTC())

	// Output:
	// [a b c] map[a:a b:b c:c] key=value true dbhost 1m0s 2007-01-02 15:04:05 +0000 UTC
}

type SomeCustomType struct {
	key   string
	value string
}

func (s *SomeCustomType) UnmarshalText(text []byte) error {
	split := strings.SplitN(string(text), "=", 2)
	for i := range split {
		split[i] = strings.TrimSpace(split[i])
	}

	s.key = split[0]
	s.value = split[1]

	return nil
}

func (s SomeCustomType) String() string {
	return fmt.Sprintf("%s=%s", s.key, s.value)
}
