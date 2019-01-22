package gmbh

import (
	"github.com/gimlet-gmbh/gimlet/gproto"
	"github.com/gimlet-gmbh/gimlet/ipc"
)

/*
 * gmbh.go
 *
 * Abe Dick
 * Nov 2018
 */

func handleDataRequest(req gproto.Request) (*gproto.Responder, error) {

	request := ipc.RequestFromProto(req)
	responder := ipc.Responder{}

	handler, ok := c.registeredFunctions[request.Method]
	if !ok {
		responder.HadError = true
		responder.ErrorString = "Could not locate method in -cabal_generic"
	} else {
		handler(request, &responder)
	}

	return responder.ToProto(), nil
}

func createLog() {

}
