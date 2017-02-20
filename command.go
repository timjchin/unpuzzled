package cli

import (
	"flag"
	"fmt"
)

type Author struct {
	Name  string
	Email string
}

type Command struct {
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
	for _, variable := range c.Variables {
		variable.setFlag(c.flagSet)
	}
	err := c.flagSet.Parse(c.args[:])

	for _, variable := range c.Variables {
		val, found := variable.getFlagValue(c.flagSet)
		fmt.Println(variable.GetName(), val, found)
	}
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

