package gmbh

/*
 * api.go
 * Abe Dick
 * Nov 2018
 */

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gimlet-gmbh/gimlet/ipc"
)

type (

	// Request -
	Request = ipc.Request

	// Responder -
	Responder = ipc.Responder

	// handlerFunc is the function type
	handlerFunc = func(req ipc.Request, resp *ipc.Responder)
)

/*
 @ Point

 About to refactor the whole API to get read a config instead of
 setting properties with a config

*/

func init() {

}

// var c *Cabal

// // Cabal - singleton for cabal coms
// type Cabal struct {
// 	name     string
// 	isServer bool
// 	isClient bool
// 	address  string

// 	registeredFunctions map[string]func(req ipc.Request, resp *ipc.Responder)
// }

// func newService()

// NewComsModule - start coms
func NewComsModule() *Cabal {
	c = &Cabal{
		isClient: false,
		isServer: false,

		registeredFunctions: map[string]handlerFunc{},
	}
	return c
}

// SetServer - Tell CORE that this service is capable of implementing
// all functionality required to be make data requests to.
func (c *Cabal) SetServer() {
	c.isServer = true
}

// SetClient - Tell CORE that this service is capable of making data
// requests to other services.
func (c *Cabal) SetClient() {
	c.isClient = true
}

// Start - Registers the service with CORE.
func (c *Cabal) Start(name string) {
	c.name = name
	addr, _ := _ephemeralRegisterService(name, c.isClient, c.isServer)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		done <- true
	}()
	if addr != "" {
		go rpcConnect(addr)
	}

	<-done
	_makeUnregisterRequest(c.name)
	os.Exit(0)
}

// Route - Callback functions to be used when handling data
// requests from CORE
func (c *Cabal) Route(route string, handler func(ipc.Request, *ipc.Responder)) {
	// *NOTE* add checking for duplicates, etc...
	c.registeredFunctions[route] = handler
}

// MakeRequest - Make a data request across Cabal
func (c *Cabal) MakeRequest(target, method, data string) ipc.Responder {
	resp, err := _makeDataRequest(target, method, data)
	if err != nil {
		panic(err)
	}
	return resp
}
