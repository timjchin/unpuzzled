package unpuzzled

import (
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type IntVariable struct {
	Name            string
	Description     string
	Default         int
	Required        bool
	Destination     *int
	flagDestination *int
}

func (i *IntVariable) GetName() string {
	return i.Name
}

func (i *IntVariable) GetDescription() string {
	return i.Description
}

func (i *IntVariable) GetDestination() interface{} {
	return i.Destination
}

func (i *IntVariable) IsRequired() bool {
	return i.Required
}

func (i *IntVariable) GetDefault() (interface{}, bool) {
	if i.Default == 0 {
		return nil, false
	} else {
		return i.Default, true
	}
}

func (i *IntVariable) apply(val interface{}) {
	switch val.(type) {
	case int:
		value := val.(int)
		*i.Destination = value
	case int64:
		value := val.(int64)
		*i.Destination = int(value)
	case float64:
		value := val.(float64)
		*i.Destination = int(value)
	}
}

func (i *IntVariable) setDefaults() {
	if i.Default != 0 {
		*i.Destination = i.Default
	}
}

func (i *IntVariable) setFlag(flagset *flag.FlagSet) {
	var destination int
	i.flagDestination = &destination
	flagset.IntVar(i.flagDestination, i.Name, i.Default, i.Description)
}

func (i *IntVariable) getFlagValue(set *flag.FlagSet) (interface{}, bool) {
	if *i.flagDestination == 0 {
		return nil, false
	} else {
		return *i.flagDestination, true
	}
}

func (i *IntVariable) setEnv(value string, envName string) (interface{}, bool) {
	int64val, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"envName":  envName,
			"err":      err,
			"envValue": value,
			"name":     i.Name,
		}).Fatal("Failed to parse from environment.")
	}
	return int(int64val), true
}
