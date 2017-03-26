package unpuzzled

import (
	"fmt"
	"testing"
	"time"

	"os"

	"github.com/stretchr/testify/assert"
)

type fullTestConfig struct {
	TestString   string
	TestFloat64  float64
	TestBool     bool
	TestInt      int
	TestInt64    int64
	TestDuration time.Duration
}

type argumentAssignemntTest struct {
	Name       string
	Command    *Command
	Arguments  []string
	Validation func(*testing.T, *Command)
}

// Test the general split of arguments between general flags, and nested subcommands.
// Each command should only get the set of arguments between the current command and the next nested subcommand.
func TestArgumentAssignment(t *testing.T) {
	tests := []argumentAssignemntTest{
		argumentAssignemntTest{
			Name: "Basic argument assignment, no subcommands.",
			Command: &Command{
				Name: "main",
			},
			Arguments: []string{"--a=5", "--b", "--c"},
			Validation: func(t *testing.T, c *Command) {
				assert.Equal(t, []string{"--a=5", "--b", "--c"}, c.args, "Single command should get all the arguments.")
			},
		},
		argumentAssignemntTest{
			Name: "Multiple subcommands, call middle command.",
			Command: &Command{
				Name: "main",
				Subcommands: []*Command{
					&Command{
						Name: "server",
						Subcommands: []*Command{
							&Command{
								Name: "metrics",
							},
						},
					},
				},
			},
			Arguments: []string{"server", "--random-value=5", "--another-value=b"},
			Validation: func(t *testing.T, c *Command) {
				assert.Equal(t, []string{"--random-value=5", "--another-value=b"}, c.Subcommands[0].args, "Middle server command should get all the arguments.")
			},
		},
		argumentAssignemntTest{
			Name: "Multiple subcommands",
			Command: &Command{
				Name: "main",
				Subcommands: []*Command{
					&Command{
						Name: "sub-a",
					},
					&Command{
						Name: "sub-b",
					},
				},
			},
			Arguments: []string{"main-a", "main-b", "sub-a", "--a", "--b", "--c"},
			Validation: func(t *testing.T, c *Command) {
				assert.Equal(t, []string{"main-a", "main-b"}, c.args, "Main command should get first two args.")
				assert.Equal(t, []string{"--a", "--b", "--c"}, c.Subcommands[0].args, "sub-a command should get the remaining args.")
				assert.Nil(t, c.Subcommands[1].args, "sub-b command should receive no arguments.")

				assert.Equal(t, true, c.Subcommands[0].Active, "sub-a should be marked as active.")
				assert.Equal(t, false, c.Subcommands[1].Active, "sub-b should be marked as inactive.")
			},
		},
		argumentAssignemntTest{
			Name: "Multiple Nested Commands",
			Command: &Command{
				Name: "main",
				Subcommands: []*Command{
					&Command{
						Name: "sub-a",
						Subcommands: []*Command{
							&Command{
								Name: "sub-b",
								Subcommands: []*Command{
									&Command{
										Name: "sub-b",
									},
								},
							},
						},
					},
				},
			},
			Arguments: []string{"main-a", "sub-a", "--a", "--b", "--c", "sub-b", "--d", "--e", "--f", "sub-b", "--g", "--h"},
			Validation: func(t *testing.T, c *Command) {
				assert.Equal(t, []string{"main-a"}, c.args, "Main command should get first two args.")
				assert.Equal(t, []string{"--a", "--b", "--c"}, c.Subcommands[0].args, "sub-a command should get the remaining args.")
				assert.Equal(t, []string{"--d", "--e", "--f"}, c.Subcommands[0].Subcommands[0].args, "sub-b command should get the remaining args.")
				assert.Equal(t, []string{"--g", "--h"}, c.Subcommands[0].Subcommands[0].Subcommands[0].args, "sub-c command should get the remaining args.")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test.Command.assignArguments(test.Arguments)
			test.Validation(t, test.Command)
		})
	}
}

type testExpandedName struct {
	Command       *Command
	ExpectedNames []string
	Name          string
}

// Internally used to bring commands / subcommands into a single namespace.
// Test to make sure these names are properly generated, should be chained together with "."'s.
func TestGetExpandedName(t *testing.T) {
	tests := []testExpandedName{
		testExpandedName{
			Name: "nested",
			Command: &Command{
				Name: "first",
				Subcommands: []*Command{
					&Command{
						Name: "second",
					},
					&Command{
						Name: "third",
					},
				},
			},
			ExpectedNames: []string{
				"first",
				"first.second",
				"first.third",
			},
		},
		testExpandedName{
			Name: "nested",
			Command: &Command{
				Name: "first",
				Subcommands: []*Command{
					&Command{
						Name: "second",
						Subcommands: []*Command{
							&Command{
								Name: "third",
							},
						},
					},
				},
			},
			ExpectedNames: []string{
				"first",
				"first.second",
				"first.second.third",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test.Command.buildTree(nil)
			var names []string
			loopCommands(test.Command, func(c *Command) {
				names = append(names, c.GetExpandedName())
			}, true)
			assert.Equal(t, test.ExpectedNames, names, "expanded names are not equal.")
		})
	}
}

func loopCommands(command *Command, fn func(*Command), isFirst bool) {
	if isFirst {
		fn(command)
	}
	for _, subcommand := range command.Subcommands {
		fn(subcommand)
		loopCommands(subcommand, fn, false)
	}
}

type testEnvironmentVariables struct {
	Name     string
	Command  *Command
	EnvVars  []envVar
	Expected *fullTestConfig
}
type envVar struct {
	Key   string
	Value string
}

func TestEnvironmentVariables(t *testing.T) {
	config := &fullTestConfig{}

	tests := []testEnvironmentVariables{
		testEnvironmentVariables{
			Name: "strings and bools",
			Command: &Command{
				Name: "envTest",
				Variables: []Variable{
					&StringVariable{
						Name:        "test-value",
						Destination: &config.TestString,
					},
					&BoolVariable{
						Name:        "test-bool",
						Destination: &config.TestBool,
					},
					&Float64Variable{
						Name:        "test-float-64",
						Destination: &config.TestFloat64,
					},
					&DurationVariable{
						Name:        "test-duration",
						Destination: &config.TestDuration,
					},
				},
			},
			Expected: &fullTestConfig{
				TestString:   "a",
				TestBool:     true,
				TestFloat64:  float64(1.5),
				TestDuration: time.Minute,
			},
			EnvVars: []envVar{
				envVar{"TEST_VALUE", "a"},
				envVar{"TEST_BOOL", "true"},
				envVar{"TEST_FLOAT_64", "1.5"},
				envVar{"TEST_DURATION", "1m"},
			},
		},
		testEnvironmentVariables{
			Name: "ensure false bools",
			Command: &Command{
				Name: "envTest",
				Variables: []Variable{
					&BoolVariable{
						Name:        "test-bool",
						Destination: &config.TestBool,
					},
					&StringVariable{
						Name:        "test-value",
						Destination: &config.TestString,
					},
					&Float64Variable{
						Name:        "test-float-64",
						Destination: &config.TestFloat64,
					},
					&Int64Variable{
						Name:        "test-int-64",
						Destination: &config.TestInt64,
					},
					&IntVariable{
						Name:        "test-int",
						Destination: &config.TestInt,
					},
					&DurationVariable{
						Name:        "test-duration",
						Destination: &config.TestDuration,
					},
				},
			},
			Expected: &fullTestConfig{
				TestString:   "b",
				TestBool:     false,
				TestFloat64:  float64(2.5),
				TestInt64:    int64(100),
				TestInt:      10,
				TestDuration: time.Second * time.Duration(15),
			},
			EnvVars: []envVar{
				envVar{"TEST_VALUE", "b"},
				envVar{"TEST_BOOL", "false"},
				envVar{"TEST_FLOAT_64", "2.5"},
				envVar{"TEST_INT_64", "100"},
				envVar{"TEST_INT", "10"},
				envVar{"TEST_DURATION", "15s"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			for _, envVar := range test.EnvVars {
				os.Setenv(envVar.Key, envVar.Value)
			}

			app := NewApp()
			app.Command = test.Command
			app.Silent = true
			app.args = []string{}
			app.parseCommands()

			assert.Equal(t, test.Expected, config, "Full test config should be the same for ENV values.")

			for _, envVar := range test.EnvVars {
				assert.NoError(t, os.Unsetenv(envVar.Key), "Should not error while unsetting the env var.")
			}
		})
	}
}

type testCLIFlags struct {
	Name     string
	Command  *Command
	Expected *fullTestConfig
	Args     []string
}

func TestCLIFlags(t *testing.T) {
	config := &fullTestConfig{}

	tests := []testCLIFlags{
		testCLIFlags{
			Name: "basic",
			Command: &Command{
				Name: "basic",
				Variables: []Variable{
					&StringVariable{
						Name:        "test-value",
						Destination: &config.TestString,
					},
					&BoolVariable{
						Name:        "test-bool",
						Destination: &config.TestBool,
					},
					&Float64Variable{
						Name:        "test-float-64",
						Destination: &config.TestFloat64,
					},
					&DurationVariable{
						Name:        "test-duration",
						Destination: &config.TestDuration,
					},
				},
			},
			Expected: &fullTestConfig{
				TestString:   "random",
				TestBool:     true,
				TestFloat64:  float64(1.5),
				TestDuration: time.Second * time.Duration(12),
			},
			Args: []string{
				"path_to_exec",
				"--test-value=random",
				"--test-bool=true",
				"--test-float-64=1.5",
				"--test-duration=12s",
			},
		},
		testCLIFlags{
			Name: "subommand",
			Command: &Command{
				Name: "basic",
				Variables: []Variable{
					&StringVariable{
						Name:        "test-value",
						Destination: &config.TestString,
					},
					&BoolVariable{
						Name:        "test-bool",
						Destination: &config.TestBool,
					},
				},
				Subcommands: []*Command{
					&Command{
						Name: "nested",
						Variables: []Variable{
							&StringVariable{
								Name:        "test-value",
								Destination: &config.TestString,
							},
							&BoolVariable{
								Name:        "test-bool",
								Destination: &config.TestBool,
							},
							&Float64Variable{
								Name:        "test-float-64",
								Destination: &config.TestFloat64,
							},
							&IntVariable{
								Name:        "test-int",
								Destination: &config.TestInt,
							},
							&DurationVariable{
								Name:        "test-duration",
								Destination: &config.TestDuration,
							},
							&Int64Variable{
								Name:        "test-int-64",
								Destination: &config.TestInt64,
							},
						},
					},
				},
			},
			Expected: &fullTestConfig{
				TestString:   "used",
				TestBool:     true,
				TestFloat64:  float64(2.5),
				TestInt:      5,
				TestInt64:    int64(100),
				TestDuration: time.Hour,
			},
			Args: []string{
				"path_to_exec",
				"--test-value=ignored",
				"--test-bool=true",
				"nested",
				"--test-value=used",
				"--test-float-64=2.5",
				"--test-int=5",
				"--test-int-64=100",
				"--test-duration=1h",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			app := NewApp()
			app.Command = test.Command
			app.Silent = true
			app.Run(test.Args)

			assert.Equal(t, test.Expected, config, "test config values should be equal.")
		})
	}
}

type testTomlConfig struct {
	Name       string
	Command    *Command
	ConfigPath string
	Expected   *fullTestConfig
}

func TestTomlConfig(t *testing.T) {
	config := &fullTestConfig{}

	tests := []testTomlConfig{
		testTomlConfig{
			Name:       "Basic",
			ConfigPath: "./fixtures/basic_test.toml",
			Command: &Command{
				Name: "basic",
				Variables: []Variable{
					&Float64Variable{
						Name:        "testfloat",
						Description: "Setting a float64 variable.",
						Destination: &config.TestFloat64,
					},
					&StringVariable{
						Name:        "teststring",
						Description: "Setting a string variable.",
						Destination: &config.TestString,
					},
					&BoolVariable{
						Name:        "testbool",
						Description: "Setting a bool variable.",
						Destination: &config.TestBool,
					},
					&IntVariable{
						Name:        "testint",
						Description: "Setting an integer variable.",
						Destination: &config.TestInt,
					},
					&Int64Variable{
						Name:        "testint64",
						Description: "Testing an int64 variable.",
						Destination: &config.TestInt64,
					},
					&ConfigVariable{
						StringVariable: &StringVariable{
							Required:    true,
							Name:        "config",
							Description: "Main configuration",
						},
						Type: TomlConfig,
					},
					&DurationVariable{
						Name:        "test-duration",
						Destination: &config.TestDuration,
					},
				},
			},
			Expected: &fullTestConfig{
				TestFloat64:  float64(1.2345),
				TestString:   "hi",
				TestBool:     true,
				TestInt:      5,
				TestInt64:    int64(100),
				TestDuration: time.Duration(15) * time.Second,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			app := NewApp()
			app.Command = test.Command
			app.Silent = true
			app.Run([]string{"path_to_exec", fmt.Sprintf("--config=%s", test.ConfigPath)})
			assert.Equal(t, test.Expected, config, "config values should be the same.")
		})
	}
}

type testJsonConfig struct {
	Name           string
	Command        *Command
	ConfigPath     string
	Expected       *fullTestConfig
	ExpectedNested *fullTestConfig
}

func TestJsonConfig(t *testing.T) {
	config := &fullTestConfig{}
	nestedConfig := &fullTestConfig{}

	tests := []testJsonConfig{
		testJsonConfig{
			Name:       "Basic",
			ConfigPath: "./fixtures/basic_test.json",
			Command: &Command{
				Name: "basic",
				Variables: []Variable{
					&Float64Variable{
						Name:        "test-float",
						Description: "Setting a float64 variable.",
						Destination: &config.TestFloat64,
					},
					&StringVariable{
						Name:        "test-string",
						Description: "Setting a string variable.",
						Destination: &config.TestString,
					},
					&BoolVariable{
						Name:        "test-bool",
						Description: "Setting a bool variable.",
						Destination: &config.TestBool,
					},
					&IntVariable{
						Name:        "test-int",
						Description: "Setting an integer variable.",
						Destination: &config.TestInt,
					},
					&Int64Variable{
						Name:        "test-int-64",
						Description: "Testing an int64 variable.",
						Destination: &config.TestInt64,
					},
					&DurationVariable{
						Name:        "test-duration",
						Destination: &config.TestDuration,
					},
					&ConfigVariable{
						StringVariable: &StringVariable{
							Required:    true,
							Name:        "config",
							Description: "Main configuration",
						},
						Type: JsonConfig,
					},
				},
				Subcommands: []*Command{
					&Command{
						Name: "nested",
						Variables: []Variable{
							&Float64Variable{
								Name:        "test-float",
								Description: "Setting a float64 variable.",
								Destination: &nestedConfig.TestFloat64,
							},
							&StringVariable{
								Name:        "test-string",
								Description: "Setting a string variable.",
								Destination: &nestedConfig.TestString,
							},
							&BoolVariable{
								Name:        "test-bool",
								Description: "Setting a bool variable.",
								Destination: &nestedConfig.TestBool,
							},
							&IntVariable{
								Name:        "test-int",
								Description: "Setting an integer variable.",
								Destination: &nestedConfig.TestInt,
							},
							&Int64Variable{
								Name:        "test-int-64",
								Description: "Testing an int64 variable.",
								Destination: &nestedConfig.TestInt64,
							},
							&DurationVariable{
								Name:        "test-duration",
								Destination: &nestedConfig.TestDuration,
							},
						},
					},
				},
			},
			Expected: &fullTestConfig{
				TestFloat64:  float64(1.2345),
				TestString:   "hi",
				TestBool:     true,
				TestInt:      5,
				TestInt64:    int64(100),
				TestDuration: time.Hour,
			},
			ExpectedNested: &fullTestConfig{
				TestFloat64:  float64(1.2345),
				TestString:   "hi",
				TestBool:     true,
				TestInt:      5,
				TestInt64:    int64(100),
				TestDuration: time.Hour,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			app := NewApp()
			app.Command = test.Command
			app.Silent = true
			app.Run([]string{"path_to_exec", fmt.Sprintf("--config=%s", test.ConfigPath), "nested"})
			assert.Equal(t, test.Expected, config, "config values should be the same.")
			assert.Equal(t, test.ExpectedNested, nestedConfig, "config values should be the same.")
		})
	}
}

type testLoopActiveCommands struct {
	Name     string
	Command  *Command
	Args     []string
	Expected string
}

// based on the arguments given, test that the loop function actually calls
// all the active commands, and not the nonactive ones.
func TestLoopActiveCommands(t *testing.T) {
	tests := []testLoopActiveCommands{
		testLoopActiveCommands{
			Name: "Single nested",
			Command: &Command{
				Name: "basic",
				Subcommands: []*Command{
					&Command{
						Name: "a",
					},
					&Command{
						Name: "b",
					},
					&Command{
						Name: "c",
					},
				},
			},
			Args:     []string{"basic", "a"},
			Expected: "basic.a.",
		},
		testLoopActiveCommands{
			Name: "Multi nested",
			Command: &Command{
				Name: "basic",
				Subcommands: []*Command{
					&Command{
						Name: "a",
						Subcommands: []*Command{
							&Command{
								Name: "a",
								Subcommands: []*Command{
									&Command{
										Name: "a",
									},
									&Command{
										Name: "b",
									},
								},
							},
							&Command{
								Name: "b",
							},
						},
					},
					&Command{
						Name: "b",
					},
				},
			},
			Args:     []string{"basic", "a", "a", "a"},
			Expected: "basic.a.a.a.",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			app := NewApp()
			app.Command = test.Command
			app.args = test.Args
			app.Silent = true
			app.parseCommands()
			outString := ""
			app.Command.loopActiveCommands(func(c *Command) {
				outString += c.Name + "."
			})
			assert.Equal(t, test.Expected, outString, "Active command loop names should be equal.")
		})
	}
}

type testDefaultValues struct {
	Name           string
	Command        *Command
	Args           []string
	ExpectedString string
	ExpectedBool   bool
}

func TestDefaultValues(t *testing.T) {
	var testString string
	var testBool bool
	tests := []testDefaultValues{
		testDefaultValues{
			Name: "Basic default test.",
			Command: &Command{
				Name: "basic",
				Variables: []Variable{
					&StringVariable{
						Name:        "test-value",
						Destination: &testString,
						Default:     "default-value",
					},
					&BoolVariable{
						Name:        "test-bool",
						Destination: &testBool,
						Default:     true,
					},
				},
			},
			Args:           []string{},
			ExpectedBool:   true,
			ExpectedString: "default-value",
		},
		testDefaultValues{
			Name: "Ensure defaults are overridden.",
			Command: &Command{
				Name: "basic",
				Variables: []Variable{
					&StringVariable{
						Name:        "test-value",
						Destination: &testString,
						Default:     "default-value",
					},
					&BoolVariable{
						Name:        "test-bool",
						Destination: &testBool,
						Default:     true,
					},
				},
			},
			Args:           []string{"--test-value=flag-value", "--test-bool=false"},
			ExpectedBool:   false,
			ExpectedString: "flag-value",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			app := NewApp()
			app.Command = test.Command
			app.Silent = true
			app.args = test.Args
			app.parseCommands()
			app.printOverrides()
			assert.Equal(t, test.ExpectedString, testString, "Default value for string must be equal.")
			assert.Equal(t, test.ExpectedBool, testBool, "Default value for string must be equal.")
		})
	}
}

type testRequiredValues struct {
	Name               string
	Command            *Command
	Args               []string
	ExpectedMissingMap map[string][]Variable
}

func TestRequiredValues(t *testing.T) {
	var stringVar string
	var boolVar bool

	requiredStringVar := &StringVariable{
		Name:        "test-value",
		Destination: &stringVar,
		Required:    true,
	}

	requiredBoolVar := &BoolVariable{
		Required:    true,
		Name:        "test-bool",
		Destination: &boolVar,
	}

	requiredStringVarWithDefault := &StringVariable{
		Name:        "test-value-b",
		Default:     "test",
		Destination: &stringVar,
		Required:    true,
	}

	tests := []testRequiredValues{
		testRequiredValues{
			Name: "Basic required test.",
			Command: &Command{
				Name:      "basic",
				Variables: []Variable{requiredBoolVar, requiredStringVar},
			},
			Args: []string{},
			ExpectedMissingMap: map[string][]Variable{
				"basic": []Variable{
					requiredBoolVar,
					requiredStringVar,
				},
			},
		},
		testRequiredValues{
			Name: "Variable that has a default, but is required will not appear in the list.",
			Command: &Command{
				Name:      "basic",
				Variables: []Variable{requiredBoolVar, requiredStringVar, requiredStringVarWithDefault},
			},
			Args: []string{},
			ExpectedMissingMap: map[string][]Variable{
				"basic": []Variable{
					requiredBoolVar,
					requiredStringVar,
				},
			},
		},
		testRequiredValues{
			Name: "Test nested values",
			Command: &Command{
				Name:      "basic",
				Variables: []Variable{requiredBoolVar, requiredStringVar, requiredStringVarWithDefault},
				Subcommands: []*Command{
					&Command{
						Name:      "basic",
						Variables: []Variable{requiredBoolVar, requiredStringVar, requiredStringVarWithDefault},
					},
				},
			},
			Args: []string{"basic"},
			ExpectedMissingMap: map[string][]Variable{
				"basic": []Variable{
					requiredBoolVar,
					requiredStringVar,
				},
				"basic.basic": []Variable{
					requiredBoolVar,
					requiredStringVar,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			app := NewApp()
			app.Command = test.Command
			app.Silent = true
			app.args = test.Args
			app.parseCommands()
			app.checkRequiredVariables()
			assert.Equal(t, test.ExpectedMissingMap, app.missingRequiredVariables, "Required variables should be the same.")
		})
	}
}
