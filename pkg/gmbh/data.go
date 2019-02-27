package gmbh

import (
	"errors"

	"github.com/gmbh-micro/rpc/intrigue"
)

/**********************************************************************************
**** Handling Data Requests
**********************************************************************************/

// HandlerFunc is the publically exposed function to register and use the callback functions
// from within gmbhCore. Its behavior is modeled after the http handler that is baked into go
// by default
type HandlerFunc = func(req Request, resp *Responder)

// Route - Callback functions to be used when handling data
// requests from gmbh or other services
//
// TODO: Add a mechanism to safely add these and check for collisions, etc.
func (g *Client) Route(route string, handler HandlerFunc) {
	g.registeredFunctions[route] = handler
}

// MakeRequest is the default method for making data requests through gmbh
func (g *Client) MakeRequest(target, method string, data *Payload) (Responder, error) {
	resp, err := makeDataRequest(target, method, data)
	if err != nil {
		return Responder{}, errors.New("could not complete request: " + err.Error())
	}
	return resp, nil
}

func handleDataRequest(req intrigue.Request) (*intrigue.Responder, error) {

	var request Request
	request = requestFromProto(&req)
	responder := Responder{}

	handler, ok := g.registeredFunctions[request.transport.Method]
	if !ok {
		g.printer("could not find hander=%s", request.transport.Method)
		responder.err = "could not find method in service map"
	} else {
		g.printer("sending to hander=%s", request.transport.Method)
		handler(request, &responder)
	}
	protoResponder := responder.proto()
	g.printer(protoResponder.String())
	return protoResponder, nil
}
