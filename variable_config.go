package cli

import (
	"errors"
	"flag"
	"io/ioutil"

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
	ErrFailedToLoadToml = errors.New("Failed to load toml.")
)

func (c *ConfigVariable) GetName() string {
	return c.StringVariable.Name
}

func (c *ConfigVariable) apply(interface{}) {
	log.Fatal("Config Variable does not apply values.")
}

func (c *ConfigVariable) ParseConfig(set *flag.FlagSet) error {
	// ignore error, library should handle this.
	stringPointer, _ := c.StringVariable.getFlagValue(set)
	value := stringPointer.(string)
	data, err := ioutil.ReadFile(value)
	if err != nil {
		return err
	}

	switch c.Type {
	case TomlConfig:
		tree, err := toml.Load(string(data))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal(ErrFailedToLoadToml)
			return ErrFailedToLoadToml
		}
		config := &tomlConfig{
			tree: tree,
		}
		c.config = config

		treeMap := tree.ToMap()
	default:
		log.Fatal("Unimplemented config")

	}
	return nil
}

func (c *ConfigVariable) getConfigValue(path string) (interface{}, error) {
	return c.config.GetByVariable(path)
}

type tomlConfig struct {
	tree *toml.TomlTree
}

func (t *tomlConfig) GetByVariable(path string) (interface{}, error) {
	return t.tree.Get(path), nil
}
