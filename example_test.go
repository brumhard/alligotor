package alligotor_test

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/brumhard/alligotor"
)

type Config struct {
	WillStayDefault string            // not overwritten in any source -> will keep default
	SomeList        []string          `description:"list that can ba assigned in format: a,b,c"` // description used for flag usage
	SomeMap         map[string]string `config:"env=MAP"`                                         // overwrite the env var to read
	SomeCustomType  SomeCustomType    `config:"env=custom"`                                      // implements text.Unmarshaler
	API             struct {
		Enabled *bool // pointers to basic types are also supported
	}
	DB struct {
		HostName string        `config:"flag=h host"` // overwrite the cli flags to read, h is shortname
		Timeout  time.Duration `config:"flag=time"`   // only overwrite long name
	}
	TimeStamp  time.Time `config:"file=custom"`                       // implements text.Unmarshaler, overwrite key in file
	Everything string    `config:"env=every,flag=e every,file=every"` // set overwrites for every source
}

func Example() {
	dir, _ := os.MkdirTemp("", "testing")
	defer os.RemoveAll(dir)

	jsonBytes := []byte(`{
    "custom": "2007-01-02T15:04:05Z",
	"db": {
		"timeout": "2m0s"
	}
}`)

	os.Args = []string{"cmdName", "--somelist", "a,b,c", "--api-enabled", "true", "-h", "dbhost", "--every", "every"}
	_ = os.Setenv("TEST_MAP", "a=a,b=b,c=c")
	_ = os.Setenv("TEST_DB_TIMEOUT", "1m0s")
	_ = os.Setenv("TEST_CUSTOM", "key=value")

	filePath := path.Join(dir, "example_config.json")
	_ = os.WriteFile(filePath, jsonBytes, os.ModePerm)

	cfg := Config{WillStayDefault: "yessir"}

	// The order of sources will also set the order in which the sources overwrite each other.
	// That's why the db timeout set in the json is overwritten with the one set in env variable.
	cfgReader := alligotor.New(
		alligotor.NewFilesSource([]string{dir}, []string{"example_config"}),
		alligotor.NewEnvSource("TEST"),
		alligotor.NewFlagsSource(),
	)

	// There's also a default reader, that can be used with alligotor.Get().
	if err := cfgReader.Get(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(
		cfg.WillStayDefault, cfg.SomeList, cfg.SomeMap, cfg.SomeCustomType,
		*cfg.API.Enabled, cfg.DB.HostName, cfg.DB.Timeout, cfg.TimeStamp.UTC(),
		cfg.Everything,
	)

	// Output:
	// yessir [a b c] map[a:a b:b c:c] key=value true dbhost 1m0s 2007-01-02 15:04:05 +0000 UTC every
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
