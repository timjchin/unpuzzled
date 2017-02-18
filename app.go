package cli

import (
	"flag"

	log "github.com/sirupsen/logrus"
)

type App struct {
	Name            string
	Usage           string
	LongDescription string
	Copyright       string
	Command         *Command
	Authors         []Author
	Action          func()
	ConfigFlag      string
	flagSet         *flag.FlagSet
	args            []string
}

func NewApp() *App {
	return &App{
		Name:    "cli",
		Authors: make([]Author, 0),
	}
}

func (a *App) Run(args []string) {
	a.flagSet = flag.NewFlagSet(a.Name, flag.ContinueOnError)
	a.args = args[1:]

	a.checkForConfig()
	a.parseCommands()

	err := a.flagSet.Parse(a.args)
	if err != nil {
		return
	}
}

func (a *App) checkForConfig() {
	// if a.ConfigFlag != "" {
	// }
}

// func (a *App) parseCommands() {
// 	for _, command := range a.Commands {
// 		for _, variable := range command.Variables {
// 			value, set := variable.setEnv()
// 			fmt.Println("setenv", value, set)
// 			variable.setFlag(a.flagSet)
// 		}
// 	}
// 	a.flagSet.Parse(a.args)
// 	for _, command := range a.Commands {
// 		for _, variable := range command.Variables {
// 			variable.getFlagValue(a.flagSet)
// 		}
// 	}
// }

func (a *App) parseCommands() {
	if a.Command == nil {
		log.Fatal("No command attached to the app!")
	}
	a.Command.buildTree(nil)
	a.Command.assignArguments(a.args)
	a.Command.parseFlags()
}
