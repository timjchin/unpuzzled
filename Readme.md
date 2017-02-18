

Goals
===
- Handles multiple types of configuration files, while making it clear to the end user which values are being chosen.
    - This should include:
        - ENV Variables
        - JSON / YAML / TOML
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

