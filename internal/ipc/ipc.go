package ipc

import (
	"github.com/gimlet-gmbh/gimlet/gproto"
)

/**
 * ipc.go
 * Abe Dick
 * January 2019
 */

// Request is the data that will be routed and sent to the requested service in which to
// gather data from
type Request struct {
	// Sender is the name of the service that is sending the message
	Sender string

	// Target is the name or alias of the intended recepient
	Target string

	// Method is the handler to invoke on target entry
	// TODO: Change this as it can be considered confusing with
	// 		 the HTTP methods...
	Method string

	// Data1 is the data to send
	// TODO: remove this and more articulately handle data
	Data1 string
}

// Responder is the data that will be sent back to the service that requested the data
type Responder struct {
	// Result is the resulting datat from target
	// TODO: remove this and more articulately handle data
	Result string

	// ErrorString is the corresponding error string if HadError is true
	ErrorString string

	// HadError is true if the request was not completed without error
	HadError bool
}

// RequestFromProto takes a gproto request and returns the corresponding
// ipc.Request object
func RequestFromProto(r gproto.Request) Request {
	return Request{
		Sender: r.Sender,
		Target: r.Target,
		Method: r.Method,
		Data1:  r.Data1,
	}
}

// ToProto returns the gproto Request object corresponding to the current
// ipc.Request object
func (r *Request) ToProto() *gproto.Request {
	return &gproto.Request{
		Sender: r.Sender,
		Target: r.Target,
		Method: r.Method,
		Data1:  r.Data1,
	}
}

// ResponderFromProto takes a gproto Responder and returns the corresponding
// ipc.Responder object
func ResponderFromProto(r gproto.Responder) Responder {
	return Responder{
		Result:      r.Result,
		ErrorString: r.ErrorString,
		HadError:    r.HadError,
	}
}

// ToProto returns the gproto Request object corresponding to the current
// ipc.Responder object
func (r *Responder) ToProto() *gproto.Responder {
	return &gproto.Responder{
		Result:      r.Result,
		ErrorString: r.ErrorString,
		HadError:    r.HadError,
	}
}
