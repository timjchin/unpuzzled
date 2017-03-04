package unpuzzled

import (
	"flag"
	"regexp"
	"strings"
)

type Variable interface {
	GetName() string
	GetDescription() string
	GetDestination() interface{}
	// return the default value, bool if it's set or not.
	GetDefault() (interface{}, bool)
	IsRequired() bool

	setDefaults()
	setFlag(*flag.FlagSet)
	setEnv() (interface{}, bool)
	apply(interface{})
	getFlagValue(*flag.FlagSet) (interface{}, bool)
}

var osToEnvReplaceRegexp = regexp.MustCompile(`[\.\-]`)

func convertNameToOS(name string) string {
	a := osToEnvReplaceRegexp.ReplaceAllString(strings.ToUpper(name), "_")
	return a
}
