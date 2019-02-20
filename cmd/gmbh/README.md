# gmbh

## Usage

### Starting a gmbh server
`gmbh --core --config=<yaml_config_path>`

Options
* `--verbose` for all output to stdout
* `--verbose-data` for all gmbh data to stdout, process management data is logged

### Starting a gmbh process manager
`gmbhPM`

### Attaching remote processes
`gmbh --container --config=<yaml_config_path>`

### Reporting data from gmbh

`gmbh --list` lists all remotes currently attached
`gmbh --list-one=<id>` lists data from one remote
`gmbh --report` lists data in report form with errors
`gmbh --restart` sends a restart signal to all remotes
`gmbh --restart-one=<id>` sends a restart signal to one remote
`gmbh -q` shuts down gmbh