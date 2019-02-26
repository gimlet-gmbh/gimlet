package gmbh

import "github.com/gmbh-micro/rpc/intrigue"

// Request is the publically exposed requester between services in gmbh
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

// ToProto returns the gproto Request object corresponding to the current
// Request object
func (r *Request) toProto() *intrigue.Request {
	return &intrigue.Request{
		Sender: r.Sender,
		Target: r.Target,
		Method: r.Method,
		Data1:  r.Data1,
	}
}
