package unpuzzled

import (
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type BoolVariable struct {
	Name        string
	Description string
	Required    bool
	Default     bool
	Destination *bool

	flagDestination *bool
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

func (b *BoolVariable) GetDestination() interface{} {
	return b.Destination
}

func (b *BoolVariable) GetDefault() (interface{}, bool) {
	if b.Default {
		return b.Default, true
	} else {
		return b.Default, false
	}
}

func (b *BoolVariable) apply(val interface{}) {
	if boolVal, ok := val.(bool); ok {
		*b.Destination = boolVal
	}
}

func (b *BoolVariable) setDefaults() {
	*b.Destination = b.Default
}

func (b *BoolVariable) setFlag(flagset *flag.FlagSet) {
	var destination bool
	b.flagDestination = &destination
	flagset.BoolVar(b.flagDestination, b.Name, b.Default, b.Description)
}

func (b *BoolVariable) getFlagValue(set *flag.FlagSet) (interface{}, bool) {
	return *b.flagDestination, true
}

func (b *BoolVariable) setEnv(value string, envName string) (interface{}, bool) {
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.WithFields(log.Fields{
			"envName":  envName,
			"envValue": value,
			"err":      err,
			"name":     b.Name,
		}).Fatal("Failed to parse bool value from environment.")
	}
	return boolValue, true
}
