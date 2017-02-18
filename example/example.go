package main

import (
	"fmt"
	"os"

	"github.com/timjchin/cli"
)

type Config struct {
	CustomString string
	CustomBool   bool
}

func main() {
	app := cli.NewApp()
	config := &Config{}

	app.ConfigFlag = "config, c"
	app.Command = &cli.Command{
		Name: "Main Command",
		Variables: []cli.Variable{
			&cli.StringVariable{
				Name:        "random-value",
				Description: "Here's a random string",
				Destination: &config.CustomString,
			},
			&cli.BoolVariable{
				Name:        "booltest",
				Description: "Bool test",
				Destination: &config.CustomBool,
			},
		},
		Action: func() {
			fmt.Println("Running main command.")
		},
		Subcommands: []*cli.Command{
			&cli.Command{
				Name: "server",
				Variables: []cli.Variable{
					&cli.StringVariable{
						Name:        "random-value",
						Description: "Here's a random string",
						Destination: &config.CustomString,
					},
				},
				Action: func() {
					fmt.Println("Running server command.")
				},
				Subcommands: []*cli.Command{
					&cli.Command{
						Name: "metrics",
						Variables: []cli.Variable{
							&cli.StringVariable{
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
