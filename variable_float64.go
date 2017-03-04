package unpuzzled

import (
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Float64Variable struct {
	Name            string
	Description     string
	Default         float64
	Required        bool
	Destination     *float64
	flagDestination *float64
}

func (f *Float64Variable) GetName() string {
	return f.Name
}

func (f *Float64Variable) GetDescription() string {
	return f.Description
}

func (f *Float64Variable) GetDestination() interface{} {
	return f.Destination
}

func (f *Float64Variable) IsRequired() bool {
	return f.Required
}

func (f *Float64Variable) GetDefault() (interface{}, bool) {
	if f.Default == float64(0) {
		return nil, false
	} else {
		return f.Default, true
	}
}

func (f *Float64Variable) apply(val interface{}) {
	switch val.(type) {
	case int:
		value := val.(int)
		*f.Destination = float64(value)
	case int64:
		value := val.(int64)
		*f.Destination = float64(value)
	case float64:
		value := val.(float64)
		*f.Destination = value
	}
}

func (f *Float64Variable) setDefaults() {
	if f.Default != float64(0) {
		*f.Destination = f.Default
	}
}

func (f *Float64Variable) setFlag(flagset *flag.FlagSet) {
	var destination float64
	f.flagDestination = &destination
	flagset.Float64Var(f.flagDestination, f.Name, f.Default, f.Description)
}

func (f *Float64Variable) getFlagValue(set *flag.FlagSet) (interface{}, bool) {
	if *f.flagDestination == float64(0) {
		return nil, false
	} else {
		return *f.flagDestination, true
	}
}

func (f *Float64Variable) setEnv(value string, envName string) (interface{}, bool) {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"envName":  envName,
			"err":      err,
			"envValue": value,
			"name":     f.Name,
		}).Fatal("Failed to parse float from environment.")
	}
	return floatVal, true
}
