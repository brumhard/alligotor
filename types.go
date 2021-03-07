package alligotor

import (
	"encoding"
	"reflect"
	"time"
)

// nolint: gochecknoglobals // package lvl type definitions
var (
	zeroString      = ""
	zeroDuration    = time.Duration(0)
	stringType      = reflect.TypeOf(zeroString)
	stringPtrType   = reflect.TypeOf(&zeroString)
	durationType    = reflect.TypeOf(zeroDuration)
	durationPtrType = reflect.TypeOf(&zeroDuration)
	timeType        = reflect.TypeOf(time.Time{})
	timePtrType     = reflect.TypeOf(&time.Time{})
	stringSliceType = reflect.TypeOf([]string{})
	stringMapType   = reflect.TypeOf(map[string]string{})
	textUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)
