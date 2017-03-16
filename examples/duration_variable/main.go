package main

import (
	"fmt"
	"os"
	"time"

	"github.com/timjchin/unpuzzled"
)

func main() {
	var duration time.Duration
	app := unpuzzled.NewApp()
	app.Command = &unpuzzled.Command{
		Name:  "main",
		Usage: "An example application that uses an unpuzzled.DurationVariable.",
		Variables: []unpuzzled.Variable{
			&unpuzzled.DurationVariable{
				Name:        "test-duration",
				Description: "An example time.Duration variable.",
				Destination: &duration,
				Default:     time.Second,
			},
		},
		Action: func() {
			fmt.Println("Duration has been parsed!", duration)
		},
	}

	app.Run(os.Args)
}
