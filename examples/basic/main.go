package main

import (
	"fmt"
	"os"
	"time"

	"github.com/timjchin/unpuzzled"
)

type Config struct {
	OtherString    string
	CustomString   string
	ThirdString    string
	CustomBool     bool
	CustomFloat64  float64
	CustomDuration time.Duration
	Int64Val       int64
}

func main() {
	app := unpuzzled.NewApp()
	config := &Config{}
	app.OverridesOutputInTable = true
	app.Copyright = "Fake Co. 2017"
	app.Authors = []unpuzzled.Author{
		unpuzzled.Author{
			Name:  "Tim",
			Email: "email@email.com",
		},
		unpuzzled.Author{
			Name:  "Another author",
			Email: "email@email.com",
		},
	}
	app.Command = &unpuzzled.Command{
		Name:  "main",
		Usage: "An example application that prints values.",
		Variables: []unpuzzled.Variable{
			&unpuzzled.StringVariable{
				Name:        "random-value",
				Description: "Here's a random string",
				Default:     "random",
				Destination: &config.OtherString,
			},
			&unpuzzled.BoolVariable{
				Name:        "test-bool-duplicate",
				Description: "An example boolean variable.",
				Destination: &config.CustomBool,
			},
			&unpuzzled.BoolVariable{
				Name:        "test-bool",
				Description: "An example required boolean variable.",
				Destination: &config.CustomBool,
				Required:    true,
			},
			&unpuzzled.Float64Variable{
				Name:        "test-float",
				Description: "An example required float variable.",
				Destination: &config.CustomFloat64,
				Required:    true,
			},
			&unpuzzled.ConfigVariable{
				StringVariable: &unpuzzled.StringVariable{
					Required:    true,
					Name:        "config",
					Description: "An example required configuration variable.",
				},
				Type: unpuzzled.TomlConfig,
			},
			&unpuzzled.DurationVariable{
				Name:        "test-duration",
				Description: "An example time.Duration variable.",
				Destination: &config.CustomDuration,
			},
		},
		Action: func() {
			fmt.Println("Running main command.")
		},
		Subcommands: []*unpuzzled.Command{
			&unpuzzled.Command{
				Name:  "server",
				Usage: "Run a server",
				Variables: []unpuzzled.Variable{
					&unpuzzled.StringVariable{
						Name:        "nested-string",
						Description: "An example required varible for a nested command.",
						Destination: &config.ThirdString,
						Required:    true,
					},
					&unpuzzled.Int64Variable{
						Name:        "nested-int-64",
						Description: "An example optional int64 value for a nested command.",
						Destination: &config.Int64Val,
					},
				},
				Action: func() {
					fmt.Println("Running server command.")
					fmt.Println(config.CustomString, config.OtherString)
					fmt.Println(config.CustomBool)
				},
				Subcommands: []*unpuzzled.Command{
					&unpuzzled.Command{
						Name: "metrics",
						Variables: []unpuzzled.Variable{
							&unpuzzled.StringVariable{
								Name:        "random-value",
								Description: "Here's a random string",
								Destination: &config.CustomString,
							},
						},
						Action: func() {
							fmt.Println("Running server metrics command.")
						},
					},
				},
			},
		},
	}

	app.Run(os.Args)

}
