package main

import (
	"fmt"
	"os"

	"github.com/timjchin/unpuzzled"
)

// This example showcases how unpuzzled will handle overwritten pointers.
// In this setup, both "example-a", and "example-b" point to `testString`.
// If both values are set, then `example-b`'s value will be the final
// value set, and will overwrite the value of `example-a`.
//
// If an overwrite occurs, it will be output as a warning before the program is run.
func main() {
	var testString string
	app := unpuzzled.NewApp()
	app.Command = &unpuzzled.Command{
		Variables: []unpuzzled.Variable{
			&unpuzzled.StringVariable{
				Name:        "example-a",
				Destination: &testString,
			},
			&unpuzzled.StringVariable{
				Name:        "example-b",
				Destination: &testString,
			},
		},
		Action: func() {
			fmt.Println(testString)
		},
	}
	app.Run(os.Args)
}
