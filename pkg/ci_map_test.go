package pkg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ciMap_Get(t *testing.T) {
	r := require.New(t)

	jsonString := []byte(`{
	"test": {
		"innertest": "arrrr",
		"innertest2": "pirate"
	},
	"test2": "idk"
}`)

	// nested
	ciMap := newCiMap()
	r.NoError(json.Unmarshal(jsonString, ciMap))
	val, ok := ciMap.Get("test" + defaultSeparator + "innertest")
	r.True(ok)
	r.Equal("arrrr", val)

	// case insensitive
	val2, ok := ciMap.Get("test" + defaultSeparator + "INNERTEST2")
	r.True(ok)
	r.Equal("pirate", val2)

	// in root
	val3, ok := ciMap.Get("TEST2")
	r.True(ok)
	r.Equal("idk", val3)
}
