# gmbh CLI
This tool starts a gmbh server. Right now it hangs onto the process and outputs debug info to stdOut.

### Usage
In the main project directory run the `gmbh` command from terminal and a gmbh server will start.

### Expected file directory for gmbh project
/
&nbsp;&nbsp;gmbh/
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{gmbh_files}
&nbsp;&nbsp;services/
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{service folders w/ gmbh config file}

gmbh_files will be where the CLI looks to start the gmbh process. The gmbh process will then scan through the services directory looking for all subdirectories that have a gmbh config YAML file with the correct instructions. 

### Config file layout
This is preliminary and will change frequently during development.
```
name: <service_name>            # The name of the service
aliases: ["<other", "names>"]   # Any aliases you wish the service to be known as
language: go                    # The language of the service (now only in Golang)
makefile: true                  # Is there a makefile in the same directory as the config file?
pathtobin: ./bin/???            # Where does the makefile output the binary to?
isClient: true                  # Does this service call other services?
isServer: true                  # Can other services call this service?
```