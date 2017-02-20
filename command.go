package cli

import (
	"flag"
	"io/ioutil"
	"strings"
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
		parentCommand   *Command
		flagSet         *flag.FlagSet
		args            []string
		Action          func()
		Active          bool
		expandedName    string
	}

	activeSetting struct {
		CommandPath  string      `json:"command_path"`
		VariableName string      `json:"variable_name"`
		Value        interface{} `json:"value"`
		Source       ParsingType `json:"source"`
	}
)

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
	fn(c)
	for _, command := range c.Subcommands {
		if !command.Active {
			continue
		}
		fn(command)
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

func (c *Command) getSetFlags() []activeSetting {
	var allSettings []activeSetting
	c.loopActiveVariables(func(command *Command, variable Variable) {
		expandedName := command.GetExpandedName()
		if val, set := variable.getFlagValue(command.flagSet); set {
			allSettings = append(allSettings, activeSetting{
				CommandPath:  expandedName,
				VariableName: variable.GetName(),
				Value:        val,
				Source:       CliFlags,
			})
		}
	})
	return allSettings
}

func (c *Command) parseEnvVars() []activeSetting {
	var allSettings []activeSetting
	c.loopActiveVariables(func(command *Command, variable Variable) {
		expandedName := command.GetExpandedName()
		if val, set := variable.setEnv(); set {
			allSettings = append(allSettings, activeSetting{
				CommandPath:  expandedName,
				VariableName: variable.GetName(),
				Value:        val,
				Source:       EnvironmentVariables,
			})
		}
	})
	return allSettings
}
