package pkg

import "testing"

func TestConfig(t *testing.T) {
	testStruct := struct {
		Port      int `config:"default=443,env=PORT,flag=port p"`
		Something string
	}{}

	config := Config{
		Files: []string{"./config"},
		Env:   false,
	}
	config.Get(&testStruct)

	// default should be set
}
