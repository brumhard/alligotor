package alligotor_test

import (
	"fmt"
	"log"

	"github.com/brumhard/alligotor"
)

func Example_defaults() {
	cfg := struct {
		Anything string
	}{
		Anything: "TheDefaultValue",
	}

	if err := alligotor.Get(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.Anything)

	// Output:
	// TheDefaultValue
}
