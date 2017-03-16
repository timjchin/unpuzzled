package unpuzzled

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
)

var zeroDuration = time.Duration(0)

type DurationVariable struct {
	Name            string
	Description     string
	Default         time.Duration
	Required        bool
	Destination     *time.Duration
	flagDestination *time.Duration
}

func (d *DurationVariable) GetName() string {
	return d.Name
}

func (d *DurationVariable) GetDescription() string {
	return d.Description
}

func (d *DurationVariable) GetDestination() interface{} {
	return d.Destination
}

func (d *DurationVariable) IsRequired() bool {
	return d.Required
}

func (d *DurationVariable) GetDefault() (interface{}, bool) {
	if d.Default == zeroDuration {
		return zeroDuration, false
	} else {
		return d.Default, true
	}
}

func (d *DurationVariable) apply(val interface{}) {
	if duration, ok := val.(time.Duration); ok {
		*d.Destination = duration
	} else if stringVal, ok := val.(string); ok {
		duration, err := time.ParseDuration(stringVal)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("Failed to parse time.Duration.")
		}
		*d.Destination = duration
	}
}

func (d *DurationVariable) setDefaults() {
	if d.Default != zeroDuration {
		*d.Destination = d.Default
	}
}
func (d *DurationVariable) setFlag(flagset *flag.FlagSet) {
	var destination time.Duration
	d.flagDestination = &destination
	flagset.DurationVar(d.flagDestination, d.Name, d.Default, d.Description)
}

func (d *DurationVariable) getFlagValue(set *flag.FlagSet) (interface{}, bool) {
	if *d.flagDestination == zeroDuration {
		return nil, false
	} else {
		return *d.flagDestination, true
	}
}

func (d *DurationVariable) setEnv(value string, envName string) (interface{}, bool) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"environmentVariable": envName,
		}).Fatal("Failed to parse time.Duration from Environment.")
		return nil, false
	}
	return duration, true
}
