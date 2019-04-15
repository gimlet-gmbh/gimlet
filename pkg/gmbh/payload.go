package gmbh

import (
	"encoding/json"
	"fmt"

	"github.com/gmbh-micro/rpc/intrigue"
)

// Payload handles data that is to be transported between services
type Payload struct {
	// JSON label->json; the object will be marshalled into JSON
	JSON map[string][]byte
}

// NewPayload returns an empty new payload
func NewPayload() *Payload {
	return &Payload{}
}

// Get returns the value of payload.JSON at key
func (p *Payload) Get(key string) interface{} {
	if p.JSON == nil {
		return nil
	}

	var obj interface{}
	v, ok := p.JSON[key]
	if !ok {
		fmt.Println("could not find in map")
		return obj
	}
	err := json.Unmarshal(v, &obj)
	if err != nil {
		fmt.Println(err)
	}
	return obj
}

// GetAsString returns the string value of payload.JSON at key
func (p *Payload) GetAsString(key string) string {
	value := p.Get(key)
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

// Append adds a value to Payload.JSON; overwrites current value as default behavior.
func (p *Payload) Append(key string, value interface{}) {
	if p.JSON == nil {
		p.JSON = make(map[string][]byte)
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	p.JSON[key] = bytes
}

// AppendDataMap adds all values of the input map to the payload.JSON
func (p *Payload) AppendDataMap(inputMap map[string][]byte) {
	if inputMap == nil {
		return
	}
	if p.JSON == nil {
		p.JSON = inputMap
	}
	for k, v := range inputMap {
		p.JSON[k] = v
	}
}

// Proto ; parses payload to protocall buffer
func (p *Payload) Proto() *intrigue.Payload {
	proto := &intrigue.Payload{}
	if p == nil {
		return proto
	}
	if p.JSON != nil {
		m := make(map[string][]byte)
		for k, v := range p.JSON {
			j, e := json.Marshal(v)
			if e != nil {
				em := "error marshalling json from Go"
				copy(m[k], em)
			}
			m[k] = j
		}
		proto.JSON = m
	}
	return proto
}

// payloadFromProto ;
func payloadFromProto(proto *intrigue.Payload) *Payload {
	p := &Payload{}
	p.JSON = proto.GetJSON()
	return p
}
