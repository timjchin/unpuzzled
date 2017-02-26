package unpuzzled

import (
	"flag"
	"os"
)

type BoolVariable struct {
	Name        string
	Description string
	Required    bool
	Default     bool
	Destination *bool

	flagDestination *bool
	envName         string
}

func (b *BoolVariable) GetName() string {
	return b.Name
}

func (b *BoolVariable) GetDescription() string {
	return b.Description
}

func (b *BoolVariable) IsRequired() bool {
	return b.Required
}

func (b *BoolVariable) apply(val interface{}) {
	if boolVal, ok := val.(bool); ok {
		*b.Destination = boolVal
	}
}

func (b *BoolVariable) setFlag(flagset *flag.FlagSet) {
	var destination bool
	b.flagDestination = &destination
	flagset.BoolVar(b.flagDestination, b.Name, b.Default, b.Description)
}

func (b *BoolVariable) getFlagValue(set *flag.FlagSet) (interface{}, bool) {
	if *b.Destination {
		return true, true
	}
	return false, false
}

func (b *BoolVariable) setEnv() (interface{}, bool) {
	b.envName = convertNameToOS(b.Name)
	if value, found := os.LookupEnv(b.envName); found {
		return value, true
	}
	return nil, false
}
