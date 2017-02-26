# Unpuzzled
A user-first CLI library. 

Many CLI tools support importing variables from different sources: command line flags, environment variables, and configuration variables. 

When your application is being used, it's often not clear what values are set, or why.

Unpuzzled gives you and your users a clear explanation of where variables are being set, and which ones are being overwritten.

Clarity prevents confusion.

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

