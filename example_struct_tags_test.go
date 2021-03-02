package alligotor_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/brumhard/alligotor"
)

type StructTagConfig struct {
	API struct {
		Port     int    `config:"env=PORT,flag=p port"`
		LogLevel string `config:"env=LOG,file=default.log"`
	}
	DB struct {
		LogLevel string `config:"env=LOG,file=default.log"`
	}
}

// Example_structTags shows how the struct tags can be used to set other names for the config sources.
// In this case the API.Port property can not only be set with the env variable PREFIX_API_PORT but also
// with just PORT. In cases where both variables are set the explicit one will have higher priority.
//
// Like this it is also possible to set default names for the properties and overwrite them in cases you need that.
// In the following example both log levels can be set from the env variable "LOG" or the value at path
// default.log (<rootField><separator><childFieldOfRootField>) in the file.
// So in general in the defaults section of the json file all log levels can be set to debug but maybe the API needs
// to be debugged so its loglevel can be explicitly set in the api object in the json.
// You could also overwrite with the PREFIX_API_LOGLEVEL environment variable.
//
// Im this example type string is used as type for loglevel, but zapcore.Level and logrus.Level are also
// supported out of the box. It's just not used here to mimize the package's dependencies.
//
// Also flags short and long name can be set in the struct tag.
func Example_structTags() {
	dir, _ := ioutil.TempDir("", "testing")
	defer os.RemoveAll(dir)

	jsonBytes := []byte(`{
    "default": {
        "log": "error"
    },
    "api": {
        "port": 1234,
        "logLevel": "debug"
    }
}`)

	filePath := path.Join(dir, "example_config.json")
	_ = ioutil.WriteFile(filePath, jsonBytes, 0600)

	os.Args = []string{"cmdName", "-p", "2345"}

	collector := alligotor.New(
		alligotor.NewFilesSource([]string{dir}, "example_config", alligotor.WithFileSeparator(".")),
		alligotor.NewEnvSource("PREFIX", alligotor.WithEnvSeparator("_")),
		alligotor.NewFlagsSource(alligotor.WithFlagSeparator(".")),
	)

	var cfg StructTagConfig
	_ = collector.Get(&cfg)

	fmt.Println(cfg)

	// Output:
	// {{2345 debug} {error}}
}
