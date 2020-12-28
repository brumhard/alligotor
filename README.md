# Alligotor

The zero configuration configuration package.

## Install

```shell script
go get github.com/brumhard/alligotor
```

## What is Alligotor?

Alligotor is designed to be used as the configuration source for executables (not commands in a command line application)
for example for apis or any other long running application that need a startup config.

It takes only a few lines of code to get going, and it supports:

- setting defaults just like you're used to from for example json unmarshalling (see this [example](example_defaults_test.go))
- reading from YAML and JSON files
- reading from environment variables
- reading from command line flags
- disabling sources
- extremely simple API
- support for every type (by implementing TextUnmarshaler) and out of the box support for many common ones
- setting paths in each configuration source for default values (see the [example](example_struct_tags_test.go))

## Why Alligotor?

There are a lot of configuration packages for Go that give you the ability to load you configuration from
several sources like env vars, command line flags or config files.

Alligotor was designed to have the least configuration effort possible while still keeping it customizable.
That's why if you keep the package defaults you only need one function call and your config struct definition
to fill this struct with values from environment variables, several config files and also command line flags.

Also default can default paths can be used to keep your configuration as small as possible.
Default paths means that you set a name/path in the configuration source from which multiple fields take their
value if they're not set in any other way.
See this [example](example_struct_tags_test.go) for further information. 

## Minimal example

```Go
package main

import (

"github.com/brumhard/alligotor"
"go.uber.org/zap/zapcore"
"time"
)

func main() {
    // define the config struct
    cfg := struct{
        SomeList []string
        SomeMap  map[string]string
        API      struct {
            Enabled  bool
            LogLevel zapcore.Level
        }
        DB struct {
            HostName string
            Timeout  time.Duration
        }
    }{
        // could define defaults here
    }
    
    // get the values
    _ = alligotor.Get(&cfg)
}
```

## TODO

- Description for fields either in struct tags or in json schema
- Support for other file formats