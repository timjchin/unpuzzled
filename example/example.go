package main

import (
	"fmt"
	"os"

	"github.com/timjchin/unpuzzled"
)

type Config struct {
	OtherString  string
	CustomString string
	CustomBool   bool
}

func main() {
	app := unpuzzled.NewApp()
	config := &Config{}

	app.Command = &unpuzzled.Command{
		Name:  "main",
		Usage: "An example application that prints values.",
		Variables: []unpuzzled.Variable{
			&unpuzzled.StringVariable{
				Name:        "random-value",
				Description: "Here's a random string",
				Destination: &config.OtherString,
			},
			&unpuzzled.BoolVariable{
				Name:        "booltest",
				Description: "Bool test",
				Destination: &config.CustomBool,
				Required:    true,
			},
			&unpuzzled.ConfigVariable{
				StringVariable: &unpuzzled.StringVariable{
					Required:    true,
					Name:        "config",
					Description: "Main configuration",
				},
				Type: unpuzzled.JsonConfig,
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
						Name:        "random-value",
						Description: "Here's a random string",
						Destination: &config.CustomString,
					},
				},
				Action: func() {
					fmt.Println("Running server command.")
					fmt.Println(config.CustomString, config.OtherString)
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

	app.Action = func() {
		fmt.Println("parsed: customstring", config.CustomString)
		fmt.Println("parsed: custombool", config.CustomBool)
	}

	app.Run(os.Args)

}
