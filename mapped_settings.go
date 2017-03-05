package unpuzzled

import (
	"fmt"
	"html/template"
	"os"
	"reflect"

	"github.com/olekukonko/tablewriter"
)

type mappedSettings struct {
	MainMap         map[string]map[string][]*activeSetting `json:"main_map"`
	OrderedSettings []*orderedSettingGroup
}

type orderedSettingGroup struct {
	CommandPath string
	Settings    [][]*activeSetting
}

func newMappedSettings() *mappedSettings {
	return &mappedSettings{
		MainMap: make(map[string]map[string][]*activeSetting),
	}
}

func (m *mappedSettings) addParsedArray(settings []*activeSetting) {
	for _, setting := range settings {
		if m.MainMap[setting.CommandPath] == nil {
			m.MainMap[setting.CommandPath] = make(map[string][]*activeSetting)
		}
		if m.MainMap[setting.CommandPath][setting.VariableName] == nil {
			m.MainMap[setting.CommandPath][setting.VariableName] = make([]*activeSetting, 0)
		}
		m.MainMap[setting.CommandPath][setting.VariableName] = append(m.MainMap[setting.CommandPath][setting.VariableName], setting)
	}
}

// Once all the settings have been generated, translate into arrays to ensure constant order
// Helps with printing variables without race conditions (unguaranteed order of maps).
func (m *mappedSettings) OrderSettings(commands []*Command) {
	var orderedSettings []*orderedSettingGroup
	m.loopCommands(commands, func(command *Command, variable Variable, settings []*activeSetting) {
		foundCurrent := false
		index := 0
		expandedName := command.GetExpandedName()
		for _, setSettings := range orderedSettings {
			if setSettings.CommandPath == expandedName {
				foundCurrent = true
			}
		}
		if !foundCurrent {
			index = len(orderedSettings)
			orderedSettings = append(orderedSettings, &orderedSettingGroup{
				CommandPath: expandedName,
				Settings:    make([][]*activeSetting, 0),
			})
		}
		orderedSettings[index].Settings = append(orderedSettings[index].Settings, settings)
	})
	m.OrderedSettings = orderedSettings
}

func (m *mappedSettings) loopCommands(commands []*Command, fn func(*Command, Variable, []*activeSetting)) {
	for _, command := range commands {
		expandedName := command.GetExpandedName()
		commandSettings := m.MainMap[expandedName]

		for _, variable := range command.Variables {
			variableSettings := commandSettings[variable.GetName()]
			if variableSettings == nil {
				continue
			}
			fn(command, variable, variableSettings)
		}
	}
}

func (m *mappedSettings) checkDuplicatePointers(commands []*Command) {
	// generate the map of pointers.
	pointerMap := make(map[interface{}][]*activeSetting)
	pointerOrder := make([]interface{}, 0)

	m.loopCommands(commands, func(command *Command, variable Variable, settings []*activeSetting) {
		for _, setting := range settings {
			if setting.Source == DefaultValue {
				continue
			}
			if pointerMap[setting.Destination] == nil {
				pointerMap[setting.Destination] = make([]*activeSetting, 0)
				pointerOrder = append(pointerOrder, setting.Destination)
			}
			pointerMap[setting.Destination] = append(pointerMap[setting.Destination], setting)
		}

		for _, pointer := range pointerOrder {
			settings := pointerMap[pointer]
			settingsLen := len(settings)

			// if more than 1 entry exists in the []*activeSetting for one pointer, it's a duplicate, and will be overwritten.
			if settingsLen < 2 {
				continue
			}
			// count the numbers of unique command + variable names
			commandVariableMap := make(map[string]int)
			for _, setting := range settings {
				path := setting.GetFullPath()
				commandVariableMap[path]++
			}
			for i, setting := range settings {
				path := setting.GetFullPath()
				// if there's more than one, then these are legitimate overrides, not
				// the same pointer across many variables.
				if commandVariableMap[path] > 1 {
					continue
				}

				if i != settingsLen-1 {
					setting.DuplicateDestination = true
				}
			}
		}
	})
}

// Helper to print duplciates in table format to Stdout.
func (m *mappedSettings) PrintDuplicates(commands []*Command) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Command", "Variable", "Source", "Value", "Type", "Status"})
	for _, commandSettings := range m.OrderedSettings {
		for _, settings := range commandSettings.Settings {
			length := len(settings)
			for i, setting := range settings {
				var status string
				if setting.DuplicateDestination {
					status = "x Overwritten Destination"
				} else if i != length-1 {
					status = "x Ignored"
				} else {
					status = "✔ Used"
				}
				row := []string{
					setting.CommandPath,
					setting.VariableName,
					ParsingTypeStringMap[setting.Source],
					fmt.Sprintf("%v", setting.Value),
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
	funcMap := getBaseFuncMap(noColor)

	funcMap["sourceString"] = func(setting *activeSetting) string {
		if setting.Source == EnvironmentVariables {
			return fmt.Sprintf("%s (%s)", ParsingTypeStringMap[setting.Source], convertNameToOS(setting.VariableName))
		} else if setting.Source == TomlConfig || setting.Source == JsonConfig {
			return fmt.Sprintf("%s (%s)", ParsingTypeStringMap[setting.Source], setting.SettingName)
		} else {
			return ParsingTypeStringMap[setting.Source]
		}
	}

	funcMap["stringify"] = func(x interface{}) string {
		return fmt.Sprintf("%v", x)
	}

	funcMap["getType"] = func(x interface{}) string {
		return reflect.TypeOf(x).String()
	}

	t.Funcs(funcMap)
	t.Parse(`{{ range $i, $allSettings := . -}}
-------------------------------------
{{ blue "Configuration:"}} {{ bold $allSettings.CommandPath }}
{{ range $key, $settings := $allSettings.Settings -}}
-------------
{{ range $j, $var := $settings -}}{{ $length := len $settings -}}
    {{ if $var.DuplicateDestination -}}
		{{ red $var.VariableName }} = {{ red (stringify $var.Value) }} ({{ getType $var.Value }})
	{{ red "ignored" }} {{ sourceString $var -}} {{ red " overwritten pointer." -}}
	{{ else if eq $length (plus1 $j) -}}
		{{ green $var.VariableName }} = {{ green (stringify $var.Value) }} ({{ getType $var.Value }})
	{{ green "set from" }} {{ sourceString $var -}}
	{{ else -}}
		{{ red $var.VariableName }} = {{ red (stringify $var.Value) }}
	{{ red "ignored" }} {{ sourceString $var -}}
	{{ end }}
{{ end -}}
{{ end }}
{{ end }}`)
	t.Execute(os.Stdout, m.OrderedSettings)
}
