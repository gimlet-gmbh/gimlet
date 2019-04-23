package gmbh

import (
	b64 "encoding/base64"
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
	if p == nil {
		return make(map[string]interface{})
	}

	if p.JSON == nil {
		return make(map[string]interface{})
	}

	var obj interface{}
	v, ok := p.JSON[key]
	if !ok {
		return obj
	}
	err := json.Unmarshal(v, &obj)
	if err != nil {
		return obj
	}
	return obj
}

// GetAsInt returns the int value of the key, else 0
func (p *Payload) GetAsInt(key string) int {
	value := p.Get(key)

	switch value.(type) {
	case int:
		return value.(int)
	case float64:
		return int(value.(float64))
	case string:
		return 0
	}
	return 0
}

// GetAsString returns the string value of payload.JSON at key, else returns the empty string
func (p *Payload) GetAsString(key string) string {
	value := p.Get(key)

	switch value.(type) {
	case int:
		return fmt.Sprintf("%d", value.(int))
	case float64:
		return fmt.Sprintf("%f", value.(float64))
	case string:
		return value.(string)
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}
	bstr, err := b64.URLEncoding.DecodeString(str)
	if err != nil {
		return str
	}

	str = string(bstr)
	if str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
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
