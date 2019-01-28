# gmbh 


## Installation
Run the install script in the scripts folder. Make sure that you have set your go path before doing this. For now be sure be be using Go version >= 1.10.


### Roadmap
* control server needs finished
    * service and pmgmt need refactored first
    * how to attach services on the fly
        * Do we enforce where they can and can't be?
        * do they have to be specified in the ini file? (yes)
            * if so, cannot attach new process
            * merits/ discussion?
            * How do we attach processes that are not local? i.e. in a Docker or on a different host?
                * Array in ini file? merits/ discussion?
* platform support
    * windows
        * unknown status
        * install scripts
    * linux
        * should be close
        * install scripts
    * macOS
        * refactor install scripts
        * where can gmbhCtrl find the control address from gmbhCore?
            * should we allow gmbhCtrl to be run only in the directory in which the project is stored unless 
* language support
    * go
        * needs refactored
    * python
    * node
    * other options w/ time permitting
        * java
        * c#
* standardization of config ini files
    * for individual services
    * for gmbhCore
* Should we refactor gmbh to start Core with **only** a path to a config file if we are starting it in daemon mode?
    * Should this then allow 
* Service permissions
    * how to enforce security? 
* custom data channels
    * enforcement at each end?
    * How to extend this into other languages?
* routing controls in routing package
    * Need some way to check and see what ports are open, or choose a new port if the "next" one is already in use
* Notify package
    * Slack integration
    * desktop notifications integration
* standard plugins
    * webserver
    * auth
    * users
* building example toys
    * text messaging w/ web socket
    * backend real time data mining in python?
* Protocall buffers needs refactored
    * clean up names/ standardization
* testing
    * heat mapping
    * profiling