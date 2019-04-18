# gmbh-micro 

gmbh-micro is an open source polyglot service framework. 

## Installation
Before installing make sure that Go version >= 1.11 is installed and the GOPATH is set. Then run the install script in the scripts folder.

## Getting Started

This repository contains gmbh CoreData, gmbh Proch, and the gmbh-GoClient. CoreData manages communication among all services via gRPC. ProcM manages all of the underlying processes.


---
# Using the gmbh-go Client

This client is used to connect services written in go to gmbh.

## Getting Started

#### 1) Configuring the client

gmbh requires several parameters to be passed into the constructor for the client.
**Required Parameters**
```go
service := gmbh.SetService(
gmbh.ServiceOptions{
    Name: "",                           // the identifier gmbh will use to refer to the service
    Aliases: []string{""},              // additional identifiers gmbh will use to refer to the service
    PeerGroups: []string{"universal"},  // the permissions group the service belongs to
})
```
**Optional Parameters**
```go
runtime := gmbh.SetRuntime(
    gmbh.RuntimeOptions{
        Blocking: true,                     // Should the client block the main thread until shutdown signal is received?
        Verbose: true,                      // Should the client print debug information to stdOut?
    })
```
#### 2) Instantiate the client
```go
package main

import "github.com/gmbh-micro/gmbh"

func main(){
    client, err := gmbh.NewClient(runtime, ...Options)
}
```

#### 3) Registering Routes
```go
client.Route("RouteName", routeHandler)
```

Handler functions have the type signature and behave similarly to the built in go http package ResponseWriter and Request objects.
```go 
func routeHandler(req gmbh.Request, resp *gmbh.Responder) 
```

#### 4) Starting the client
```go
    client.Start()
```

## Making Requests
```go
payload := gmbh.NewPayload()
payload.Append("<dataName>", <dataValue>)
result, err := client.MakeRequest("<serviceName>", "<registeredRoute>", payload)
```
<dataValue> is an interface{} in the Go Client.

## Handling Requests
```go
func handleData(req gmbh.Request, resp *gmbh.Responder) {

    // Get the data out of the request
    data := req.GetPayload()
    value := data.Get("<dataName>")             // Returns an interface{}
    value := data.GetAsString("<dataName2>")    // Returns data as a string

    // Create a response 
    payload := gmbh.NewPayload()
    payload.Append("<dataName>", <data>)
    
    // send the response
    resp.SetPayload(payload)
}
```
