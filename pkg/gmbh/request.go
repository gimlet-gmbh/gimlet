package gmbh

import "github.com/gmbh-micro/rpc/intrigue"

// Request is the publically exposed requester between services in gmbh
type Request struct {

	// transport holds data that is needed to route a data request
	transport *Transport

	// payload holds that data that is to be transported between services
	payload *Payload
}

// NewRequest returns a new request object initialized with the recepient information
func NewRequest(t *Transport, p *Payload) *Request {
	return &Request{
		transport: t,
		payload:   p,
	}
}

// GetPayload processes the payload request of the service
func (r *Request) GetPayload() *Payload {
	if r.payload == nil {
		return &Payload{}
	}
	return r.payload
}

// SetPayload for the request
func (r *Request) SetPayload(p *Payload) {
	r.payload = p
}

// GetTransport returns the transport details of the request
func (r *Request) GetTransport() *Transport {
	if r.transport == nil {
		return &Transport{}
	}
	return r.transport
}

// SetTransport sets the transport information
func (r *Request) SetTransport(t *Transport) {
	r.transport.Target = t.Target
	r.transport.Method = t.Method
}

func (r *Request) proto() *intrigue.Request {
	ir := &intrigue.Request{}
	if r.transport != nil {
		ir.Tport = r.transport.proto()
	}
	if r.payload != nil {
		ir.Pload = r.payload.Proto()
	}
	return ir
}

func requestFromProto(r *intrigue.Request) Request {
	return Request{
		transport: transportFromProto(r.GetTport()),
		payload:   payloadFromProto(r.GetPload()),
	}
}

// Transport handles data regarding the endpoints of a Request and MUST be
// defined before being sent
type Transport struct {
	// sender is the name of the service that is sending the message
	sender string

	// Target is the (name||alias) of the service that is the intended recepient
	Target string

	// Method is the method to invoke in the target
	Method string

	// History of the
	history []string
}

// proto ;
func (t *Transport) proto() *intrigue.Transport {
	if t == nil {
		return &intrigue.Transport{}
	}
	return &intrigue.Transport{
		Sender: t.sender,
		Target: t.Target,
		Method: t.Method,
	}
}

func transportFromProto(t *intrigue.Transport) *Transport {
	if t == nil {
		return &Transport{}
	}
	return &Transport{
		sender: t.GetSender(),
		Target: t.GetTarget(),
		Method: t.GetMethod(),
	}
}
