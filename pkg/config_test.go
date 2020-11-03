package pkg

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	r := require.New(t)

	// options:
	// - file=x overwrites default path to variable in file
	// - env=x overwrites default environment variable name to be used
	// - flag=x xy sets the flag names that should be used
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

	// read config from yaml file
	tempDir := t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.yml", []byte("port: 4000"), 0600))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Port)

	// read config from json file, case insensitive
	tempDir = t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.json", []byte("{\"PORT\": 4000}"), 0600))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Port)

	// read env, overwrites files
	r.NoError(os.Setenv("PORT", "5000"))
	r.NoError(config.Get(&testStruct))
	r.Equal(5000, testStruct.Port)
}

type testType struct {
	S []string
}

func (t *testType) UnmarshalText(text []byte) error {
	t.S = strings.Split(string(text), ",")
	return nil
}

func TestCollector_Env_UnmarshalText(t *testing.T) {
	r := require.New(t)

	// options:
	// - file=x overwrites default path to variable in file
	// - env=x overwrites default environment variable name to be used
	// - flag=x xy sets the flag names that should be used
	testStruct := struct {
		SomeThing testType `config:"env=SOMETHING"`
	}{
		SomeThing: testType{},
	}

	config := Collector{
		Files: ConfigFiles{},
		Env:   true,
		Flags: false,
	}

	r.NoError(os.Setenv("SOMETHING", "l,t,l"))
	r.NoError(config.Get(&testStruct))
	r.Equal([]string{"l", "t", "l"}, testStruct.SomeThing.S)
}

func TestCollector_FileOverwrite(t *testing.T) {
	r := require.New(t)

	// options:
	// - file=x overwrites default path to variable in file
	// - env=x overwrites default environment variable name to be used
	// - flag=x xy sets the flag names that should be used
	testStruct := struct {
		Port      int `config:"file=OVERWRITE,env=PORT,flag=port p"`
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

	// read config from yaml file
	tempDir := t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.yml", []byte("overwrite: 4000"), 0600))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Port)
}

func TestCollector_Get_Flags(t *testing.T) {
	r := require.New(t)

	// options:
	// - file=x overwrites default path to variable in file
	// - env=x overwrites default environment variable name to be used
	// - flag=x xy sets the flag names that should be used
	testStruct := struct {
		Port int `config:"flag=port p"`
	}{}

	config := Collector{
		Flags: true,
	}

	// short name flag
	os.Args = []string{"commandName", "-p", "1234"}

	r.NoError(config.Get(&testStruct))
	r.Equal(1234, testStruct.Port)

	// full name flag
	os.Args = []string{"commandName", "--port", "4567"}

	r.NoError(config.Get(&testStruct))
	r.Equal(4567, testStruct.Port)
}
