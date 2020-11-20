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
		Files: FilesConfig{
			Locations: []string{},
			BaseName:  "config",
		},
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

	// read flag, overwrites files
	os.Args = []string{"commandName", "-p", "1234"}
	r.NoError(config.Get(&testStruct))
	r.Equal(1234, testStruct.Port)
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

	testStruct := struct {
		SomeThing testType `config:"env=SOMETHING"`
	}{
		SomeThing: testType{},
	}

	config := Collector{}

	r.NoError(os.Setenv("SOMETHING", "l,t,l"))
	r.NoError(config.Get(&testStruct))
	r.Equal([]string{"l", "t", "l"}, testStruct.SomeThing.S)
}

func TestCollector_File_Overwrite(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Port      int `config:"file=OVERWRITE"`
		Something string
	}{
		Port: 8080,
	}

	config := Collector{
		Files: FilesConfig{
			Locations: []string{},
			BaseName:  "config",
		},
	}

	// read config from yaml file
	tempDir := t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.yml", []byte("overwrite: 4000"), 0600))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Port)
}

func TestCollector_File_Nested(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Sub struct {
			Port int
		}
	}{}

	config := Collector{
		Files: FilesConfig{
			BaseName: "config",
		},
	}

	// read config from json file
	tempDir := t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.json", []byte(`{"port": 4000}`), 0600))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.NotEqual(4000, testStruct.Sub.Port)

	// nested
	r.NoError(ioutil.WriteFile(tempDir+"/config.json", []byte(`{"sub": {"port": 4000}}`), 0600))
	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Sub.Port)
}

func TestCollector_File_Nested_Overwrite(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Sub struct {
			Port int `config:"file=PORT"`
		}
	}{}

	config := Collector{
		Files: FilesConfig{
			BaseName: "config",
		},
	}

	// read config from json file
	tempDir := t.TempDir()
	r.NoError(ioutil.WriteFile(tempDir+"/config.json", []byte(`{"port": 4000}`), 0600))
	config.Files.Locations = []string{tempDir}

	r.NoError(config.Get(&testStruct))
	r.Equal(4000, testStruct.Sub.Port)

	// nested
	r.NoError(ioutil.WriteFile(tempDir+"/config.json", []byte(`{"sub": {"port": 4000}}`), 0600))
	r.NoError(config.Get(&testStruct))
	r.NotEqual(4000, testStruct.Sub.Port)
}

func TestCollector_Get_Env(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Port int
	}{}

	config := Collector{}

	// short name flag
	// read env, overwrites files
	r.NoError(os.Setenv("PORT", "5000"))
	r.NoError(config.Get(&testStruct))
	r.Equal(5000, testStruct.Port)
}

func TestCollector_Get_Env_Nested(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Sub struct {
			Port int
		}
	}{}

	config := Collector{}

	// short name flag
	// read env, overwrites files
	r.NoError(os.Setenv("PORT", "5000"))
	r.NoError(config.Get(&testStruct))
	r.NotEqual(5000, testStruct.Sub.Port)

	r.NoError(os.Setenv("SUB_PORT", "5000"))
	r.NoError(config.Get(&testStruct))
	r.Equal(5000, testStruct.Sub.Port)
}

func TestCollector_Get_Env_Overwrite(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Port int `config:"env=PORT"`
	}{}

	config := Collector{}

	// short name flag
	// read env, overwrites files
	r.NoError(os.Setenv("PORT", "5000"))
	r.NoError(config.Get(&testStruct))
	r.Equal(5000, testStruct.Port)
}

func TestCollector_Get_Flags(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Port int
	}{}

	config := Collector{}

	// short name flag
	os.Args = []string{"commandName", "-p", "1234"}

	r.Error(config.Get(&testStruct))

	// full name flag
	os.Args = []string{"commandName", "--port", "4567"}

	r.NoError(config.Get(&testStruct))
	r.Equal(4567, testStruct.Port)
}

func TestCollector_Get_Nested(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Sub struct {
			Port int
		}
	}{}

	config := Collector{}

	// short name flag
	os.Args = []string{"commandName", "-p", "1234"}

	r.Error(config.Get(&testStruct))

	// full name flag
	os.Args = []string{"commandName", "--port", "4567"}

	r.Error(config.Get(&testStruct))

	// nested full name
	os.Args = []string{"commandName", "--sub.port", "8910"}

	r.NoError(config.Get(&testStruct))
	r.Equal(8910, testStruct.Sub.Port)
}

func TestCollector_Get_Flags_Overwrite(t *testing.T) {
	r := require.New(t)

	testStruct := struct {
		Port int `config:"flag=whaaaat w"`
	}{}

	config := Collector{}

	// default name flag
	os.Args = []string{"commandName", "--port", "4567"}

	r.Error(config.Get(&testStruct))

	// short name flag
	os.Args = []string{"commandName", "-w", "1234"}

	r.NoError(config.Get(&testStruct))
	r.Equal(1234, testStruct.Port)

	// full name flag
	os.Args = []string{"commandName", "--whaaaat", "4567"}

	r.NoError(config.Get(&testStruct))
	r.Equal(4567, testStruct.Port)
}
