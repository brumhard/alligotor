package pkg

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig(t *testing.T) {
	testStruct := struct {
		Port      int `config:"default=443,env=PORT,flag=port p"`
		Something string
	}{}

	config := Collector{
		File: ConfigFiles{
			Locations: []string{"./"},
			BaseName:  "config",
		},
		Env: false,
	}
	require.NoError(t, config.Get(&testStruct))

	// default should be set
}
