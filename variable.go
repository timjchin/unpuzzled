package unpuzzled

import (
	"flag"
	"regexp"
	"strings"
)

type Variable interface {
	// Get the Name
	GetName() string
	// Get the Description
	GetDescription() string
	// Get the Destination pointer
	GetDestination() interface{}
	// return the default value, bool if it's set or not.
	GetDefault() (interface{}, bool)
	// Get if it's a required variable or not
	IsRequired() bool

	setDefaults()
	setFlag(*flag.FlagSet)
	// os value, env name, return value and a bool representing if it's a valid setting
	setEnv(string, string) (interface{}, bool)
	apply(interface{})
	getFlagValue(*flag.FlagSet) (interface{}, bool)
}

var osToEnvReplaceRegexp = regexp.MustCompile(`[\.\-]`)

func convertNameToOS(name string) string {
	a := osToEnvReplaceRegexp.ReplaceAllString(strings.ToUpper(name), "_")
	return a
}
