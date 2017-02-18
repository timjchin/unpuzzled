package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/urfave/cli.v1"

	"gopkg.in/urfave/cli.v1/altsrc"
)

func main() {
	app := cli.NewApp()

	var (
		testInt64    int64
		testUint64   uint64
		testUint     uint
		testInt      int
		testInt2     int
		testDuration time.Duration
	)

	app.Flags = []cli.Flag{
		altsrc.NewInt64Flag(cli.Int64Flag{
			Name:        "test-int-64",
			Destination: &testInt64,
		}),
		altsrc.NewUint64Flag(cli.Uint64Flag{
			Name:        "test-uint",
			Destination: &testUint64,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:        "test-int",
			Destination: &testInt,
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:        "test-duration",
			Destination: &testDuration,
		}),
		cli.StringFlag{
			Name: "config, c",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name: "sub",
			Flags: []cli.Flag{cli.IntFlag{
				Name:        "test-int2",
				Destination: &testInt,
			}},
			Action: func(c *cli.Context) {
				fmt.Println(testInt)
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		fmt.Println(testInt)
		fmt.Println(testInt2)
		fmt.Println(testInt64)
		fmt.Println(testUint64)
		fmt.Println(testUint)
		fmt.Println(testDuration)
		return nil
	}

	// // TOML:
	// app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewTomlSourceFromFlagFunc("config"))

	// YAML:
	// app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))
	app.Run(os.Args)
}
