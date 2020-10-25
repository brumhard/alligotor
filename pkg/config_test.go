package pkg

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	r := require.New(t)

	// options:
	// - file=x overwrites default path to variable in file
	// - env=x overwrites default environment variable name to be used
	// - flay=x xy sets the flag names that should be used
	testStruct := struct {
		Port      int `config:"file=PORT,env=PORT,flag=port p"`
		Something string
	}{
		Port: 8080,
	}

	config := Collector{
		Files: ConfigFiles{
			Locations: []string{},
			BaseName:  "config",
		},
		Env:   false,
		Flags: false,
	}

	// initial values are kept if not overwritten
	r.NoError(config.Get(&testStruct))
	r.Equal(8080, testStruct.Port)

	// use defaults
	r.Equal("ARRRR", testStruct.Something)

	// read config from yaml file
	tempDir := t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.yml", []byte("port: 4000"), 0644))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Port)

	// read config from json file, case insensitive
	tempDir = t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.json", []byte("{\"PORT\": 4000}"), 0644))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Port)
}
