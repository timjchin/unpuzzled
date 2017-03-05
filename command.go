package unpuzzled

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type (
	Author struct {
		Name  string
		Email string
	}

	Command struct {
		Name            string
		Usage           string
		LongDescription string
		BeforeFunc      func(c *Command) error
		Subcommands     []*Command
		Variables       []Variable
		Action          func()
		Active          bool

		parentCommand *Command
		flagSet       *flag.FlagSet
		args          []string
		expandedName  string
		configVars    []*ConfigVariable
	}

	activeSetting struct {
		CommandPath          string      `json:"command_path"`
		VariableName         string      `json:"variable_name"`
		Value                interface{} `json:"value"`
		Destination          interface{} `json:"destination"`
		Source               ParsingType `json:"source"`
		SettingName          string      `json:"setting_name"`
		DuplicateDestination bool        `json:"duplicate_destination"`
	}
)

func (a *activeSetting) GetFullPath() string {
	return fmt.Sprintf("%s.%s", a.CommandPath, a.VariableName)
}

// Get the expanded name of a command, which includes the name of the parent commands, separated by a "."
// ex. main.sub1.sub2
func (c *Command) GetExpandedName() string {
	if c.expandedName != "" {
		return c.expandedName
	}

	parent := c
	names := make([]string, 0)

	for parent != nil {
		names = append(names, parent.Name)
		parent = parent.parentCommand
	}
	reverseStringSlice(names)
	c.expandedName = strings.Join(names, ".")
	return c.expandedName
}

// Get the set of active commands. All commands in the chain of arguments are considered active.
func (c *Command) GetActiveCommands() []*Command {
	var commands []*Command
	c.loopActiveCommands(func(command *Command) {
		commands = append(commands, command)
	})
	return commands
}

// Helper to get a map of expanded names to the actual command.
func (c *Command) GetExpandedActiveCommmands() map[string]*Command {
	outMap := make(map[string]*Command)
	c.loopActiveCommands(func(command *Command) {
		outMap[command.GetExpandedName()] = command
	})
	return outMap
}

// Helper to get a map of variables by variable name.
func (c *Command) GetVariableMap() map[string]Variable {
	outMap := make(map[string]Variable)
	for _, variable := range c.Variables {
		if _, exists := outMap[variable.GetName()]; exists {
			log.WithFields(log.Fields{
				"variable": variable.GetName(),
				"command":  c.Name,
			}).Fatal("Duplciate variables seen with the same name.")
		}
		outMap[variable.GetName()] = variable
	}
	return outMap
}

// Adds the parentCommands to all nested commands.
func (c *Command) buildTree(parentCommand *Command) {
	if parentCommand != nil {
		c.parentCommand = parentCommand
	}
	if c.Subcommands != nil {
		for _, subCommand := range c.Subcommands {
			subCommand.buildTree(c)
		}
	}
}

// Split the arguments between the main command (global arguments), and arguments for each nested subcommand.
// ex: go run main.go [global flags] subcommand [subcommand arguments] another-subcommand [another-subcommand arguments]
func (c *Command) assignArguments(args []string) *Command {
	if c.Subcommands == nil {
		c.args = args[:]
		c.Active = true
		return c
	}

	foundArgument := false

commandLoop:
	for _, command := range c.Subcommands {
		for i, arg := range args {
			if arg == command.Name {
				c.args = args[:i]
				command.assignArguments(args[i+1:])
				foundArgument = true
				c.Active = true
				break commandLoop
			}
		}
	}

	if !foundArgument {
		c.args = args[:]
		c.Active = true
	}

	return nil
}

// set all variables to default values.
func (c *Command) applyDefaultValues() {
	c.loopActiveVariables(func(c *Command, variable Variable) {
		variable.setDefaults()
	})
}

func (c *Command) parseFlags() error {
	if c.Subcommands != nil {
		for _, command := range c.Subcommands {
			if !command.Active {
				continue
			}
			if err := command.parseFlags(); err != nil {
				return err
			}
		}
	}

	if c.args == nil {
		return nil
	}

	c.flagSet = flag.NewFlagSet(c.Name, flag.ContinueOnError)
	c.flagSet.SetOutput(ioutil.Discard)

	for _, variable := range c.Variables {
		variable.setFlag(c.flagSet)
	}
	err := c.flagSet.Parse(c.args[:])

	if err != nil {
		return err
	}
	return nil
}

// test for config variables, add command state.
func (c *Command) findConfigVars() {
	c.loopActiveVariables(func(command *Command, variable Variable) {
		if config, ok := variable.(*ConfigVariable); ok {
			command.configVars = append(command.configVars, config)
		}
	})
}

// Helper to get the active config variables by type.
func (c *Command) getConfigVarsByType(configType ParsingType) []*ConfigVariable {
	var vars []*ConfigVariable
	for _, configVar := range c.configVars {
		if configVar.Type == configType {
			vars = append(vars, configVar)
		}
	}
	return vars
}

func (c *Command) parseConfigVars() error {
	c.loopActiveCommands(func(command *Command) {
		if len(command.configVars) == 0 {
			return
		}
		for _, config := range command.configVars {
			config.ParseConfig(c.flagSet)
		}
	})
	return nil
}

// Helper function to search down the tree of commands and discover if it's a help command.
// If it is, return the active command.
func (c *Command) isHelpCommand(helpMap map[string]bool) (*Command, bool) {
	for _, arg := range c.args {
		if helpMap[arg] {
			return c, true
		}
	}

	for _, command := range c.Subcommands {
		if command.Active == false {
			continue
		}
		if helpCommand, isHelp := command.isHelpCommand(helpMap); isHelp {
			return helpCommand, true
		}
	}

	return nil, false
}

// Helper to loop through all active commands.
func (c *Command) loopActiveCommands(fn func(*Command)) {
	if c.parentCommand == nil {
		fn(c)
	}
	for _, command := range c.Subcommands {
		if !command.Active {
			continue
		}
		fn(command)
		command.loopActiveCommands(fn)
	}
}

// Helper to loop through all active command's variables.
func (c *Command) loopActiveVariables(fn func(*Command, Variable)) {
	c.loopActiveCommands(func(command *Command) {
		for _, variable := range command.Variables {
			fn(command, variable)
		}
	})
}

// loop through all active variables (including variables from subcommands),
// set from CLI flags. Return all the values that have been set.
func (c *Command) getSetFlags() []*activeSetting {
	var allSettings []*activeSetting
	c.loopActiveCommands(func(command *Command) {
		expandedName := command.GetExpandedName()
		varMap := command.GetVariableMap()

		command.flagSet.Visit(func(f *flag.Flag) {
			if variable, ok := varMap[f.Name]; ok {
				if val, set := variable.getFlagValue(command.flagSet); set {
					allSettings = append(allSettings, &activeSetting{
						CommandPath:  expandedName,
						VariableName: variable.GetName(),
						Value:        val,
						Source:       CliFlags,
						Destination:  variable.GetDestination(),
					})
				}
			}
		})
	})
	return allSettings
}

func (c *Command) getDefaultValues() []*activeSetting {
	var allSettings []*activeSetting
	c.loopActiveVariables(func(command *Command, variable Variable) {
		expandedName := command.GetExpandedName()
		if value, isSet := variable.GetDefault(); isSet {
			allSettings = append(allSettings, &activeSetting{
				CommandPath:  expandedName,
				VariableName: variable.GetName(),
				Value:        value,
				Source:       DefaultValue,
				Destination:  variable.GetDestination(),
			})
		}
	})
	return allSettings
}

// loop through all active variables (including variables from subcommands),
// set from ENV vars. Return all the values that have been set.
func (c *Command) parseEnvVars() []*activeSetting {
	var allSettings []*activeSetting
	c.loopActiveVariables(func(command *Command, variable Variable) {
		expandedName := command.GetExpandedName()
		envName := convertNameToOS(variable.GetName())
		if value, found := os.LookupEnv(envName); found {
			if val, set := variable.setEnv(value, envName); set {
				allSettings = append(allSettings, &activeSetting{
					CommandPath:  expandedName,
					VariableName: variable.GetName(),
					Value:        val,
					Source:       EnvironmentVariables,
					Destination:  variable.GetDestination(),
				})
			}
		}
	})
	return allSettings
}

func (c *Command) parseConfigValues(configVars []*ConfigVariable) []*activeSetting {
	var allSettings []*activeSetting
	c.loopActiveVariables(func(command *Command, variable Variable) {
		expandedName := command.GetExpandedName()
		for _, configVar := range configVars {
			currPath := fmt.Sprintf("%s.%s", expandedName, variable.GetName())
			value, err := configVar.getConfigValue(currPath)
			if err != nil {
				log.WithFields(log.Fields{
					"variable": variable.GetName(),
					"config":   configVar.GetName(),
				}).Fatal("Failed while parsing config variable.")
			}
			if value != nil {
				allSettings = append(allSettings, &activeSetting{
					CommandPath:  expandedName,
					VariableName: variable.GetName(),
					Value:        value,
					Source:       configVar.Type,
					SettingName:  configVar.GetName(),
					Destination:  variable.GetDestination(),
				})
			}
		}
	})
	return allSettings
}
