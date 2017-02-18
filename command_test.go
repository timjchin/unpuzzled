package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
