package alligotor_test

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/brumhard/alligotor"
)

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

func Example_unmarshalText() {
	cfg := struct {
		Custom SomeCustomType
	}{}

	_ = os.Setenv("CUSTOM", "key=value")

	if err := alligotor.Get(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.Custom)

	// Output:
	// key=value
}
