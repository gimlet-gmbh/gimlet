
### Roadmap

- [ ] Control Cabal are getting divorced
    * Control
        * Become a standalone process manager
        * Core is the brain w/ remote servers that host processes
    * Cabal
        * handles all data requests, service routing
    * gmbhCtrl
        * interacts with Control server as a client to control processes
    * gmbh
        * Key Functions
            * launching project
                * Development mode
                    * Add/ attach services
                * Delpoyment mode
                    * Only services defined in docker-compose
            * launching remote services
            * generating docker-compose for cluster
- [ ] Protobuffers need refactored badly!
- [ ] Module Support
    * No more GOPATH, hopefully makes installation easier for everybody
- [ ] Language Support 
    - [ ] Go
    - [ ] Python
    - [ ] Node
- [ ] Platform Support
    - [ ] Linux
    - [ ] MacOS
    - [ ] Windows
- [ ] Continue standardizing config files
    * Need to parameterize parameters, a large config file is good if we can make most of the options defaults.
    * **For Deployment mode, need to set a flag that reads things like core and cabal addresses from environment variables**
- [ ] Docker Support
    * Testing, testing, testing
- [ ] Move away from Makefile
- [ ] Custom data channels
    * How to enforce it at both ends?
    * Interface w/ enforcement functions
- [ ] How to choose port numbers?
- [ ] Notify package
    * Slack integration
    * Desktop integrations
    * Other communications
- [ ] Standard plugins
    * Auth
    * Webserver
    * users
- [ ] Building Toys
    * text messaging with web socket
    * Blogging platform using standard plugins
- [ ] Testing
    * Heap mapping
    * profiling