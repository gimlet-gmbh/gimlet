# gmbh 

gmbh-micro is an open source polyglot service framework.


## Installation
Before installing make sure that Go version >= 1.11 is installed and the GOPATH is set. Then run the install script in the scripts folder.


#### Development Notes

* Make sure that any realtive paths inside of a service are relative to ***the directory that the gmbh service config file is in***. That is the service launcher starts services as if the binary is running from the directory where the gmbh service config file is located