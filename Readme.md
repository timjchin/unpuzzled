# Unpuzzled
A user-first CLI library. 

With the ability to import variables from many different sources: command line flags, environment variables, and configuration variables, it's often not clear what values are set, or why.

Unpuzzled gives you and your users a clear explanation of where variables are being set, and which ones are being overwritten.

Clarity prevents confusion.

# Features
* First class importing from:
    * Environment Variables
    * JSON files
    * TOML files
    * CLI Flags
* Ability to choose the order of precendece (ex. cli flags > JSON > TOML > ENV)
* Main Command and Subcommands
* Defaults to verbose output out of the box. 
     * Ability to turn it off.
* Destination variables
* Warnings on overrides 
    * Warnings on overrides by order of precedence. 
        * If a variable is set as an ENV variable, and a CLI flag overwrites it, let the user know.
        * If two or more variables have references to the same pointer, let the user know.
* Displays what variables are set, and what they are set to.
* Add default values for variables.
* Ability to set Variables as Required.
    * If a value isn't set, print a warning to stdout, and exit.
    * If a variable has a `Default` value, it can never be marked as required, because a valid value will be set.

Left to do:
* More variable types


# How to use JSON / Toml configs:
TOML:
```
app := unpuzzled.NewApp()
app.Command = &unpuzzled.Command{
    Name: "main",
    Variables: []unpuzzled.Variable{
        &unpuzzled.ConfigVariable{
            StringVariable: &unpuzzled.StringVariable{
                Name: "config"
                Description: "Main configuration, use with `go run main.go --config=path_to_file.toml`",
                Type: unpuzzled.TomlConfig,
            },
        },
    },
}
```
JSON Config Example:
Usage:
` go run main.go --config=path_to_file.json`

```
app := unpuzzled.NewApp()
app.Command = &unpuzzled.Command{
    Name: "main",
    Variables: []unpuzzled.Variable{
        &unpuzzled.ConfigVariable{
            StringVariable: &unpuzzled.StringVariable{
                Name: "config"
                Description: "Main configuration flag, use with `go run main.go --config=path_to_file.json`",
                Type: unpuzzled.JsonConfig,
            },
        },
    },
}
```


Status:
Alpha.