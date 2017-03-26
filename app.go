package unpuzzled

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

type App struct {
	// Used to name the App in the help text.
	Name string
	// The "Usage" section in the help text.
	Usage string
	// The text used for the copyright section in the help text.
	Copyright string
	// The order in which variable sources will be parsed, values later in the array will be parsed afterwards, overwriting earlier sources.
	// Default order is: CLI Flag > Toml Config > JSON Config > Environment
	ParsingOrder []ParsingType
	// Main command attached to the app.
	Command *Command
	// List of authors used in the help text.
	Authors []Author
	// The keys in this map are used to check for help values, defaults are "help", "-h", and "--help"
	HelpCommands map[string]bool
	// If help text variables will be displayed in a table. Defaults to true.
	HelpTextVariablesInTable bool
	// If overridden variables will be displayed in a table. Defaults to false.
	OverridesOutputInTable bool
	// All output will not include color
	RemoveColor bool
	// Turn off all output
	Silent                   bool
	args                     []string
	activeCommands           []*Command
	missingRequiredVariables map[string][]Variable
	settingsMap              *mappedSettings
}

type ParsingType int

const (
	EnvironmentVariables ParsingType = iota
	JsonConfig
	TomlConfig
	CliFlags
	DefaultValue
)

var ParsingTypeStringMap = map[ParsingType]string{
	EnvironmentVariables: "Environment",
	JsonConfig:           "JSON Config",
	TomlConfig:           "Toml Config",
	CliFlags:             "CLI Flag",
	DefaultValue:         "Default Value",
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
			TomlConfig,
			CliFlags,
		},
		HelpTextVariablesInTable: true,
	}
}

// Run the app. Should be called with:
// app := cli.NewApp()
// app.Run(os.Args)
func (a *App) Run(args []string) {
	if len(args) < 1 {
		log.Fatal("Arguments must be at least 1, please run with app.Run(os.Args).")
	}
	a.args = args[1:]
	a.parseCommands()
	if a.checkRequiredVariables(); a.missingRequiredVariables != nil {
		a.PrintMissingRequiredVariables()
		os.Exit(1)
	}
	a.printOverrides()

	finalCommand := a.activeCommands[len(a.activeCommands)-1]
	if finalCommand.Action != nil {
		finalCommand.Action()
	}
}

func (a *App) parseCommands() {
	if a.Command == nil {
		log.Fatal("No command attached to the app!")
	}
	a.Command.buildTree(nil)
	a.Command.assignArguments(a.args)
	a.activeCommands = a.Command.GetActiveCommands()
	a.Command.findConfigVars()

	if helpCommand, isHelp := a.Command.isHelpCommand(a.HelpCommands); isHelp {
		a.PrintHelpCommand(helpCommand)
		os.Exit(0)
	}

	if err := a.Command.parseFlags(); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("error parsing flags.")
		return
	}

	a.Command.parseConfigVars()
	a.Command.applyDefaultValues()
	a.parseByOrder()
	a.applySettingsMap()
	a.settingsMap.checkDuplicatePointers(a.activeCommands)
}

func (a *App) checkRequiredVariables() {
	missingRequiredVariables := make(map[string][]Variable)
	a.Command.loopActiveVariables(func(c *Command, variable Variable) {
		path := c.GetExpandedName()
		variableSettingsMap := a.settingsMap.MainMap[path]
		if variable.IsRequired() && len(variableSettingsMap[variable.GetName()]) == 0 {
			if missingRequiredVariables[path] == nil {
				missingRequiredVariables[path] = make([]Variable, 0)
			}
			missingRequiredVariables[path] = append(missingRequiredVariables[path], variable)
		}
	})
	if len(missingRequiredVariables) > 0 {
		a.missingRequiredVariables = missingRequiredVariables
	}
}

func (a *App) printOverrides() {
	if a.Silent {
		return
	}
	a.settingsMap.OrderSettings(a.activeCommands)
	if a.OverridesOutputInTable {
		a.settingsMap.PrintDuplicates(a.activeCommands)
	} else {
		a.settingsMap.PrintDuplicatesStdout(a.RemoveColor)
	}
}

func (a *App) PrintMissingRequiredVariables() {
	if a.Silent {
		return
	}
	if a.missingRequiredVariables == nil {
		panic("There are no missing required variables.")
	}
	t := template.New("required-variables")
	funcMap := getBaseFuncMap(a.RemoveColor)
	t.Funcs(funcMap)
	t.Parse(`---------------------------
{{ bold (red "Missing Required Variables:") }}
---------------------------
{{ range $k, $variables := . }}
{{ blue "Command" }} : {{ $k }}
{{ range $i, $var := $variables -}}
{{ green "--"}}{{ green $var.GetName }} : {{ printf "%v" $var.Description }}
{{ end -}}
{{ end }}
`)
	t.Execute(os.Stdout, a.missingRequiredVariables)
}

type helpStruct struct {
	App          *App
	HelpCommand  *Command
	ParsingOrder []string
	UseTable     bool
}

func (a *App) PrintHelpCommand(command *Command) {
	t := template.New("required-variables")
	funcMap := getBaseFuncMap(a.RemoveColor)
	funcMap["sourceString"] = func(p ParsingType) string {
		return ParsingTypeStringMap[p]
	}
	funcMap["variableTable"] = func(command *Command) string {
		buffer := new(bytes.Buffer)
		table := tablewriter.NewWriter(buffer)
		table.SetHeader([]string{
			"Flag",
			"Default",
			"Required",
			"Env Name",
			"Description",
		})
		for _, variable := range command.Variables {
			defaultValue := "--"
			if varDefault, set := variable.GetDefault(); set {
				defaultValue = fmt.Sprintf("%v", varDefault)
			}
			required := "No"
			if variable.IsRequired() {
				required = "Required"
			}
			row := []string{
				"--" + variable.GetName(),
				defaultValue,
				required,
				convertNameToOS(variable.GetName()),
				variable.GetDescription(),
			}
			table.Append(row)
		}
		table.Render()
		return buffer.String()
	}
	t.Funcs(funcMap)
	t.Parse(`{{ bold (green "APP:") }} 
{{ .App.Name }}

{{ bold (green "COMMAND:") }} 
{{ .HelpCommand.Name }}

{{ if gt (len .HelpCommand.Usage) 0 }}{{ bold (green "COMMAND USAGE:") }}
{{ .HelpCommand.Usage }}
{{ end }}
{{ bold (green "AVAILABLE SUBCOMMANDS:")}}
{{ range $i, $c := .HelpCommand.Subcommands -}}
	{{ if eq (len $c.Usage) 0 -}}
{{ bold $c.Name }}
{{ else -}}
{{ bold $c.Name }} : {{ $c.Usage }}
{{ end -}}
{{ end -}}
{{ bold "help" }} : Print this help message

{{ bold (green "PARSING ORDER:")}} (set values will override in this order)
{{ $length := len .ParsingOrder -}}
{{ range $i, $p := .ParsingOrder -}}
	{{ if eq $length (plus1 $i) -}}
		{{ $p }}
	{{ else -}}
		{{ $p }} {{ noEscape "> " -}} 
	{{ end -}}
{{ end }}
{{ bold (green "VARIABLES:")}}
{{ if .UseTable -}}
{{ noEscape (variableTable .HelpCommand) }}
{{ else -}}
{{ range $i, $v := .HelpCommand.Variables -}}
{{ blue "--"}}{{ blue $v.GetName }} {{ if $v.IsRequired }}({{ red "Required" }}) {{ end }}{{ noEscape $v.GetDescription }}
{{ end -}}
{{ end -}}
{{ if gt (len .App.Copyright) 0 }}{{ bold (green "Copyright:") }}
{{ .App.Copyright}}
{{ end }}
{{ if gt (len .App.Authors) 0 }}{{ bold (green "Authors:") }}
{{ range $i, $author := .App.Authors -}}
{{ if gt (len $author.Name) 0 }}{{ $author.Name }}{{ end }} ({{ if gt (len $author.Email) 0 }}{{ $author.Email }}{{ end }})
{{ end -}}
{{ end }}`)

	parsingOrder := []string{}
	for _, val := range a.ParsingOrder {
		parsingOrder = append(parsingOrder, ParsingTypeStringMap[val])
	}
	reverseStringSlice(parsingOrder)
	t.Execute(os.Stdout, &helpStruct{
		App:          a,
		HelpCommand:  command,
		ParsingOrder: parsingOrder,
		UseTable:     a.HelpTextVariablesInTable,
	})
}

// use the set Parsing order to apply the variables in place, adding it to the settings map.
// The last entries in the settingsMap are the selected variables.
func (a *App) parseByOrder() {
	settingsMap := newMappedSettings()
	if a.ParsingOrder == nil {
		log.Fatal("No parsing order! Use unpuzzled.NewApp when creating an application.")
	}

	settingsMap.addParsedArray(a.Command.getDefaultValues())

	for _, order := range a.ParsingOrder {
		switch order {
		case EnvironmentVariables:
			settingsMap.addParsedArray(a.Command.parseEnvVars())

		case JsonConfig:
			vars := a.Command.getConfigVarsByType(JsonConfig)
			if len(vars) == 0 {
				continue
			}
			setValues := a.Command.parseConfigValues(vars)
			settingsMap.addParsedArray(setValues)

		case TomlConfig:
			vars := a.Command.getConfigVarsByType(TomlConfig)
			if len(vars) == 0 {
				continue
			}
			setValues := a.Command.parseConfigValues(vars)
			settingsMap.addParsedArray(setValues)

		case CliFlags:
			settingsMap.addParsedArray(a.Command.getSetFlags())
		}
	}
	a.settingsMap = settingsMap
}

func (a *App) applySettingsMap() {
	commandMap := a.Command.GetExpandedActiveCommmands()
	// loop through commands, ensure that the order of settings are constantly applied,
	// instead of looping through MainMap, which is not a consistent order.
	for _, command := range a.activeCommands {
		path := command.GetExpandedName()
		variableSettingsMap := a.settingsMap.MainMap[path]
		currCommand := commandMap[path]
		variableMap := currCommand.GetVariableMap()

		for _, variable := range command.Variables {
			name := variable.GetName()
			setting := variableSettingsMap[name]
			if setting == nil {
				continue
			}
			activeSetting := setting[len(setting)-1]
			currVariable := variableMap[activeSetting.VariableName]
			// special case, ignore a config variable.
			if _, ok := currVariable.(*ConfigVariable); ok {
				continue
			}
			currVariable.apply(activeSetting.Value)
		}
	}
}
