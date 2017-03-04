package unpuzzled

import (
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Int64Variable struct {
	Name            string
	Description     string
	Default         int64
	Required        bool
	Destination     *int64
	flagDestination *int64
}

func (i *Int64Variable) GetName() string {
	return i.Name
}

func (i *Int64Variable) GetDescription() string {
	return i.Description
}

func (i *Int64Variable) GetDestination() interface{} {
	return i.Destination
}

func (i *Int64Variable) IsRequired() bool {
	return i.Required
}

func (i *Int64Variable) GetDefault() (interface{}, bool) {
	if i.Default == 0 {
		return nil, false
	} else {
		return i.Default, true
	}
}

func (i *Int64Variable) apply(val interface{}) {
	switch val.(type) {
	case int:
		value := val.(int)
		*i.Destination = int64(value)
	case int64:
		value := val.(int64)
		*i.Destination = value
	case float64:
		value := val.(float64)
		*i.Destination = int64(value)
	}
}

func (i *Int64Variable) setDefaults() {
	if i.Default != int64(0) {
		*i.Destination = i.Default
	}
}

func (i *Int64Variable) setFlag(flagset *flag.FlagSet) {
	var destination int64
	i.flagDestination = &destination
	flagset.Int64Var(i.flagDestination, i.Name, i.Default, i.Description)
}

func (i *Int64Variable) getFlagValue(set *flag.FlagSet) (interface{}, bool) {
	if *i.flagDestination == int64(0) {
		return nil, false
	} else {
		return *i.flagDestination, true
	}
}

func (i *Int64Variable) setEnv(value string, envName string) (interface{}, bool) {
	parsedValue, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"envName":  envName,
			"err":      err,
			"envValue": value,
			"name":     i.Name,
		}).Fatal("Failed to parse from environment.")
	}
	return parsedValue, true
}
