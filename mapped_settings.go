package unpuzzled

import (
	"fmt"
	"html/template"
	"os"

	"reflect"

	"github.com/olekukonko/tablewriter"
)

type mappedSettings struct {
	MainMap map[string]map[string][]*activeSetting `json:"main_map"`
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

func (m *mappedSettings) checkDuplicatePointers() {
	// generate the map of pointers.
	pointerMap := make(map[interface{}][]*activeSetting)
	for _, commandName := range m.MainMap {
		for _, settings := range commandName {
			for _, setting := range settings {
				if setting.Source == DefaultValue {
					continue
				}
				if pointerMap[setting.Destination] == nil {
					pointerMap[setting.Destination] = make([]*activeSetting, 0)
				}
				pointerMap[setting.Destination] = append(pointerMap[setting.Destination], setting)
			}
		}
	}

	// if more than 1 entry exists in the []*activeSetting for one pointer, it's a duplicate, and will be overwritten.
	for _, settings := range pointerMap {
		settingsLen := len(settings)
		if settingsLen < 2 {
			continue
		}
		for i, setting := range settings {
			if i != settingsLen-1 {
				setting.DuplicateDestination = true
			}
		}
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
				if setting.DuplicateDestination {
					status = "x Overwritten Destination"
				} else if i != length-1 {
					status = "x Ignored"
				} else {
					status = "✔ Used"
				}
				row := []string{
					expandedName,
					setting.VariableName,
					ParingTypeStringMap[setting.Source],
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
	funcMap := getColorFuncMap(noColor)

	funcMap["sourceString"] = func(setting *activeSetting) string {
		if setting.Source == EnvironmentVariables {
			return fmt.Sprintf("%s (%s)", ParingTypeStringMap[setting.Source], convertNameToOS(setting.VariableName))
		} else if setting.Source == TomlConfig || setting.Source == JsonConfig {
			return fmt.Sprintf("%s (%s)", ParingTypeStringMap[setting.Source], setting.SettingName)
		} else {
			return ParingTypeStringMap[setting.Source]
		}
	}

	funcMap["plus1"] = func(x int) int {
		return x + 1
	}

	funcMap["stringify"] = func(x interface{}) string {
		return fmt.Sprintf("%v", x)
	}

	t.Funcs(funcMap)
	t.Parse(`{{ range $command, $variables := . -}}
-------------------------------------
{{ blue "Configuration:"}} {{ bold $command }}
{{ range $key, $vars := $variables -}}
-------------
{{ range $k, $var := $vars }}{{ $length := len $vars -}}
    {{ if $var.DuplicateDestination -}}
		{{ red $key }} = {{ red (stringify $var.Value) }}
	{{ red "ignored" }} {{ sourceString $var -}} {{ red " overwritten pointer." }}
	{{ else if eq $length (plus1 $k) -}}
		{{ green $key }} = {{ green (stringify $var.Value) }}
	{{ green "set from" }} {{ sourceString $var -}} 
	{{ else -}} 
		{{ red $key }} = {{ red (stringify $var.Value) }}
	{{ red "ignored" }} {{ sourceString $var -}} 
	{{ end }}
{{ end -}}
{{ end }}
{{ end }}`)
	t.Execute(os.Stdout, m.MainMap)
}
