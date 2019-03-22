package gmbh

import "github.com/gmbh-micro/rpc/intrigue"

// Responder is the publically exposed responder between services in gmbh
type Responder struct {

	// the resulting data
	payload *Payload

	// the meta information as set by the client
	transport *Transport

	// Errors as reported by the client during data calculation
	err string
}

// proto returns the gproto Request object corresponding to the current
// Responder object
func (r *Responder) proto() *intrigue.Responder {
	return &intrigue.Responder{
		Pload: r.payload.Proto(),
		Tport: r.transport.proto(),
		Err:   r.err,
	}
}

// SetPayload for the request
func (r *Responder) SetPayload(p *Payload) {
	r.payload = p
}

// GetError ;
func (r *Responder) GetError() string {
	return r.err
}

// GetPayload from responder
func (r *Responder) GetPayload() *Payload {
	return r.payload
}

// ResponderFromProto takes a gproto Responder and returns the corresponding
// Responder object
func responderFromProto(r intrigue.Responder) Responder {

	ret := Responder{
		err: r.GetErr(),
	}

	if r.Pload != nil {
		ret.SetPayload(payloadFromProto(r.GetPload()))
	}

	if r.Tport != nil {
		ret.transport = transportFromProto(r.GetTport())
	}

	return ret
}
