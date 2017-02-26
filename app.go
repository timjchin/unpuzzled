package cli

import (
	"fmt"
	"html/template"
	"os"

	"io/ioutil"

	"reflect"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Name            string
	Usage           string
	LongDescription string
	Copyright       string
	ParsingOrder    []ParsingType
	Command         *Command
	Authors         []Author
	HelpCommands    map[string]bool
	Action          func()
	ConfigFlag      string
	RemoveColor     bool
	args            []string
}

type ParsingType int

const (
	EnvironmentVariables ParsingType = iota
	JsonConfig
	YamlConfig
	TomlConfig
	CliFlags
)

var ParingTypeStringMap = map[ParsingType]string{
	EnvironmentVariables: "Environment",
	JsonConfig:           "JSON Config",
	YamlConfig:           "YAML Config",
	TomlConfig:           "Toml Config",
	CliFlags:             "CLI Flag",
}

// Create a new application with default values set.
func NewApp() *App {
	return &App{
		Name:    "cli",
		Authors: make([]Author, 0),
		HelpCommands: map[string]bool{
			"--help": true,
			"-h":     true,
			"help":   true,
		},
		ParsingOrder: []ParsingType{
			EnvironmentVariables,
			JsonConfig,
			YamlConfig,
			TomlConfig,
			CliFlags,
		},
	}
}

// Run the app. Should be called with:
// app := cli.NewApp()
// app.Run(os.Args)
func (a *App) Run(args []string) {
	a.args = args[1:]

	a.checkForConfig()
	a.parseCommands()

	err := a.flagSet.Parse(a.args)
	if err != nil {
		return
	}
}

func (a *App) checkForConfig() {

}

func (a *App) parseCommands() {
	if a.Command == nil {
		log.Fatal("No command attached to the app!")
	}
	a.Command.buildTree(nil)
	a.Command.assignArguments(a.args)

	if helpCommand, isHelp := a.Command.isHelpCommand(a.HelpCommands); isHelp {
		fmt.Println("help.", helpCommand.Name)
		return
	}

	if err := a.Command.parseFlags(); err != nil {
		// TODO: Fix me.
		return
	}

	a.parseByOrder()
}

func (a *App) parseByOrder() error {
	settingsMap := newMappedSettings()

	for _, order := range a.ParsingOrder {
		switch order {
		case EnvironmentVariables:
			settingsMap.addArray(a.Command.parseEnvVars())

		case JsonConfig:

		case YamlConfig:

		case TomlConfig:

		case CliFlags:
			settingsMap.addArray(a.Command.getSetFlags())
		}
	}
	settingsMap.PrintDuplicates(a.Command.GetActiveCommands())
	settingsMap.PrintDuplicatesStdout(a.RemoveColor)
	return nil
}

type mappedSettings struct {
	MainMap map[string]map[string][]activeSetting `json:"main_map"`
}

func newMappedSettings() *mappedSettings {
	return &mappedSettings{
		MainMap: make(map[string]map[string][]activeSetting),
	}
}

func (m *mappedSettings) addArray(settings []activeSetting) {
	for _, setting := range settings {
		if m.MainMap[setting.CommandPath] == nil {
			m.MainMap[setting.CommandPath] = make(map[string][]activeSetting)
		}
		if m.MainMap[setting.CommandPath][setting.VariableName] == nil {
			m.MainMap[setting.CommandPath][setting.VariableName] = make([]activeSetting, 0)
		}
		m.MainMap[setting.CommandPath][setting.VariableName] = append(m.MainMap[setting.CommandPath][setting.VariableName], setting)
	}
}

// Helper to print duplciates in table format to Stdout.
func (m *mappedSettings) PrintDuplicates(commands []*Command) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Command", "Variable", "Source", "Value", "Type", "Status"})
	for _, command := range commands {
		expandedName := command.GetExpandedName()
		if m.MainMap[expandedName] == nil {
			continue
		}
		for _, settings := range m.MainMap[expandedName] {
			length := len(settings)
			for i, setting := range settings {
				var status string
				if i == length-1 {
					status = "✔ Used"
				} else {
					status = "x Ignored"
				}
				row := []string{
					expandedName,
					setting.VariableName,
					ParingTypeStringMap[setting.Source],
					fmt.Sprintf("%s", setting.Value),
					reflect.TypeOf(setting.Value).String(),
					status,
				}
				if setting.Source == EnvironmentVariables {
					row[1] += " (" + convertNameToOS(setting.VariableName) + ")"
				}
				table.Append(row)
			}
		}
	}
	table.Render()
}

// Use a custom formatted string to print duplicates on Stdout.
func (m *mappedSettings) PrintDuplicatesStdout(noColor bool) {
	t := template.New("duplicates")
	funcMap := template.FuncMap{
		"blue":  color.BlueString,
		"red":   color.RedString,
		"green": color.GreenString,
		"bold":  color.New(color.Bold).Sprint,
		"sourceString": func(p ParsingType) string {
			return ParingTypeStringMap[p]
		},
		"plus1": func(x int) int {
			return x + 1
		},
	}
	if noColor {
		funcMap["blue"] = identityString
		funcMap["red"] = identityString
		funcMap["green"] = identityString
		funcMap["bold"] = identityString
	}
	t.Funcs(funcMap)
	t.Parse(`{{ range $command, $variables := . -}}
{{ blue "Configuration:"}} {{ bold $command }}
	{{ range $key, $vars := $variables -}}
	{{ range $k, $var := $vars }}{{ $length := len $vars -}}
		{{ if eq $length (plus1 $k) -}}
			{{ green $key }} = {{ green $var.Value }}
		{{ green "set from" }} {{ sourceString $var.Source -}} 
		{{ else -}} 
			{{ red $key }} = {{ red $var.Value }}
		{{ red "ignored from" }} {{ sourceString $var.Source -}} 
		{{ end }}
	{{ end -}}
	{{ end }}
{{ end }}`)
	t.Execute(os.Stdout, m.MainMap)
}

func identityString(s string) string {
	return s
}
