# Unpuzzled
A user-first CLI library. 

Many CLI tools support importing variables from different sources: command line flags, environment variables, and configuration variables. 

When your application is being used, it's often not clear what values are set, or why.

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
* Help text
* More variable types

Goals
===
- Handles multiple types of configuration files, while making it clear to the end user which values are being chosen.
    - This should include:
        - ENV Variables
        - JSON / TOML
        - CLI flags
    - The order of parsing should be configurable, and clear to the end user.
        - Default order is: CLI > JSON > YAML > TOML > ENV 
- Commands, Subcommands
- Required / Non required flags
- Customizable Help Text
- Access to as much data as goes in. 
- Destination Variables
- Before Func
- Clear order of parsing order and overrides
- Warnings on overrides
- Default Config Outputs
- YAML / TOML Parsings
- Order parsing doesn't matter

