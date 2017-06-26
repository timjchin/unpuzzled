package unpuzzled

import (
	"errors"
	"flag"
	"io/ioutil"

	"github.com/Jeffail/gabs"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

type ConfigVariable struct {
	*StringVariable
	Type   ParsingType
	config configGetter
}

type configGetter interface {
	GetByVariable(string) (interface{}, error)
}

var (
	ErrFailedToLoadToml  = errors.New("Failed to load toml.")
	ErrFailedToLoadJson  = errors.New("Failed to load json.")
	ErrConfigValueNotSet = errors.New("Config variable is not set.")
)

func (c *ConfigVariable) apply(interface{}) {
	log.Fatal("Config Variable does not apply values.")
}

func (c *ConfigVariable) ParseConfig(set *flag.FlagSet) error {
	// ignore error, library should handle this.
	stringPointer, _ := c.StringVariable.getFlagValue(set)
	if stringPointer == nil {
		return ErrConfigValueNotSet
	}
	value := stringPointer.(string)
	data, err := ioutil.ReadFile(value)
	if err != nil {
		return err
	}

	switch c.Type {
	case TomlConfig:
		tree, err := toml.Load(string(data))
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal(ErrFailedToLoadToml)
			return ErrFailedToLoadToml
		}
		config := &tomlConfig{
			tree: tree,
		}
		c.config = config
	case JsonConfig:
		container, err := gabs.ParseJSON(data)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal(ErrFailedToLoadJson)
			return ErrFailedToLoadJson
		}
		config := &jsonConfig{
			container: container,
		}
		c.config = config

	default:
		log.Fatal("Unimplemented config type")

	}
	return nil
}

func (c *ConfigVariable) getConfigValue(path string) (interface{}, error) {
	if c.config != nil {
		return c.config.GetByVariable(path)
	}
	return nil, nil
}

type tomlConfig struct {
	tree *toml.Tree
}

func (t *tomlConfig) GetByVariable(path string) (interface{}, error) {
	return t.tree.Get(path), nil
}

type jsonConfig struct {
	container *gabs.Container
}

func (j *jsonConfig) GetByVariable(path string) (interface{}, error) {
	return j.container.Path(path).Data(), nil
}
