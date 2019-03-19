package gmbh

import (
	"encoding/json"

	"github.com/gmbh-micro/rpc/intrigue"
)

// Payload handles data that is to be transported between services
type Payload struct {

	/*
		Larger data fields
		-> for moving large bits of data
	*/

	// Fields are things like json except for underneath can be treated as string arrays.
	// Add it with a label and then retrieve it using the lookup like a normal map
	Fields map[string][]string

	// JSON label->json; the object will be marshalled into JSON
	JSON map[string]interface{}

	/*
		Smaller data fields
		-> for moving smaller bits of data
	*/

	TextFields   map[string]string
	BoolFields   map[string]bool
	ByteFields   map[string][]byte
	IntFields    map[string]int
	Int64Fields  map[string]int64
	UintFields   map[string]uint
	Uint64Fields map[string]uint64
	DoubleFields map[string]float64
	FloatFields  map[string]float32
}

// NewPayload returns an empty new payload
func NewPayload() *Payload {
	return &Payload{
		// Fields:       make(map[string][]string),
		// JSON:         make(map[string]interface{}),
		// TextFields:   make(map[string]string),
		// BoolFields:   make(map[string]bool),
		// ByteFields:   make(map[string][]byte),
		// IntFields:    make(map[string]int),
		// Int64Fields:  make(map[string]int64),
		// UintFields:   make(map[string]uint),
		// Uint64Fields: make(map[string]uint64),
		// DoubleFields: make(map[string]float64),
		// FloatFields:  make(map[string]float32),
	}
}

// GetStringField returns the value of payload.TextFields at key
func (p *Payload) GetStringField(key string) string {
	if p == nil {
		return ""
	}
	if p.TextFields == nil {
		return ""
	}
	return p.TextFields[key]
}

// GetStringFields returns the map from payload.TextFields
func (p *Payload) GetStringFields() map[string]string {
	if p.TextFields == nil {
		return make(map[string]string)
	}
	return p.TextFields
}

// GetBoolField returns the value of payload.BoolFields at key
func (p *Payload) GetBoolField(key string) bool {
	if p.BoolFields == nil {
		return false
	}
	return p.BoolFields[key]
}

// GetBoolFields returns the map from payload.BoolFields
func (p *Payload) GetBoolFields() map[string]bool {
	if p.BoolFields == nil {
		return make(map[string]bool)
	}
	return p.BoolFields
}

// GetByteField returns the value of payload.ByteFields at key
func (p *Payload) GetByteField(key string) []byte {
	if p.ByteFields == nil {
		return []byte{}
	}
	return p.ByteFields[key]
}

// GetByteFields returns the map from payload.ByteFields
func (p *Payload) GetByteFields() map[string][]byte {
	if p.ByteFields == nil {
		return make(map[string][]byte)
	}
	return p.ByteFields
}

// GetIntField returns the value of payload.IntFields at key
func (p *Payload) GetIntField(key string) int {
	if p.IntFields == nil {
		return 0
	}
	return p.IntFields[key]
}

// GetIntFields returns the map from payload.IntFields
func (p *Payload) GetIntFields() map[string]int {
	if p.IntFields == nil {
		return make(map[string]int)
	}
	return p.IntFields
}

// GetInt64Field returns the value of payload.Int64Fields at key
func (p *Payload) GetInt64Field(key string) int64 {
	if p.Int64Fields == nil {
		return 0
	}
	return p.Int64Fields[key]
}

// GetInt64Fields returns the map from payload.Int64Fields
func (p *Payload) GetInt64Fields() map[string]int64 {
	if p.Int64Fields == nil {
		return make(map[string]int64)
	}
	return p.Int64Fields
}

// GetUintField returns the value of payload.UintFields at key
func (p *Payload) GetUintField(key string) uint {
	if p.UintFields == nil {
		return 0
	}
	return p.UintFields[key]
}

// GetUintFields returns the map from payload.UintFields
func (p *Payload) GetUintFields() map[string]uint {
	if p.UintFields == nil {
		return make(map[string]uint)
	}
	return p.UintFields
}

// GetUint64Field returns the value of payload.Uint64Fields at key
func (p *Payload) GetUint64Field(key string) uint64 {
	if p.Uint64Fields == nil {
		return 0
	}
	return p.Uint64Fields[key]
}

// GetUint64Fields returns the map from payload.Uint64Fields
func (p *Payload) GetUint64Fields() map[string]uint64 {
	if p.Uint64Fields == nil {
		return make(map[string]uint64)
	}
	return p.Uint64Fields
}

// AppendStringField adds a value to Payload.TextFields; overwrites current value as default behavior.
func (p *Payload) AppendStringField(key, value string) {
	if p.TextFields == nil {
		p.TextFields = make(map[string]string)
	}
	p.TextFields[key] = value
}

// AppendStringFields adds all values of the input map to the payload.TextFields
func (p *Payload) AppendStringFields(inputMap map[string]string) {
	if inputMap == nil {
		return
	}
	if p.TextFields == nil {
		p.TextFields = inputMap
	}
	for k, v := range inputMap {
		p.TextFields[k] = v
	}
}

// AppendBoolField adds a value to Payload.BoolFields; overwrites current value as default behavior.
func (p *Payload) AppendBoolField(key string, value bool) {
	if p.BoolFields == nil {
		p.BoolFields = make(map[string]bool)
	}
	p.BoolFields[key] = value
}

// AppendBoolFields adds all values of the input map to the payload.BoolFields
func (p *Payload) AppendBoolFields(inputMap map[string]bool) {
	if inputMap == nil {
		return
	}
	if p.BoolFields == nil {
		p.BoolFields = inputMap
	}
	for k, v := range inputMap {
		p.BoolFields[k] = v
	}
}

// AppendByteField adds a value to Payload.ByteFields; overwrites current value as default behavior.
func (p *Payload) AppendByteField(key string, value []byte) {
	if p.ByteFields == nil {
		p.ByteFields = make(map[string][]byte)
	}
	p.ByteFields[key] = value
}

// AppendByteFields adds all values of the input map to the payload.ByteFields
func (p *Payload) AppendByteFields(inputMap map[string][]byte) {
	if inputMap == nil {
		return
	}
	if p.ByteFields == nil {
		p.ByteFields = inputMap
	}
	for k, v := range inputMap {
		p.ByteFields[k] = v
	}
}

// AppendIntField adds a value to Payload.IntFields; overwrites current value as default behavior.
func (p *Payload) AppendIntField(key string, value int) {
	if p.IntFields == nil {
		p.IntFields = make(map[string]int)
	}
	p.IntFields[key] = value
}

// AppendIntFields adds all values of the input map to the payload.IntFields
func (p *Payload) AppendIntFields(inputMap map[string]int) {
	if inputMap == nil {
		return
	}
	if p.IntFields == nil {
		p.IntFields = inputMap
	}
	for k, v := range inputMap {
		p.IntFields[k] = v
	}
}

// AppendInt64Field adds a value to Payload.Int64Fields; overwrites current value as default behavior.
func (p *Payload) AppendInt64Field(key string, value int64) {
	if p.Int64Fields == nil {
		p.Int64Fields = make(map[string]int64)
	}
	p.Int64Fields[key] = value
}

// AppendInt64Fields adds all values of the input map to the payload.IntFields
func (p *Payload) AppendInt64Fields(inputMap map[string]int64) {
	if inputMap == nil {
		return
	}
	if p.Int64Fields == nil {
		p.Int64Fields = inputMap
	}
	for k, v := range inputMap {
		p.Int64Fields[k] = v
	}
}

// AppendUintField adds a value to Payload.UintFields; overwrites current value as default behavior.
func (p *Payload) AppendUintField(key string, value uint) {
	if p.UintFields == nil {
		p.UintFields = make(map[string]uint)
	}
	p.UintFields[key] = value
}

// AppendUintFields adds all values of the input map to the payload.UintFields
func (p *Payload) AppendUintFields(inputMap map[string]uint) {
	if inputMap == nil {
		return
	}
	if p.UintFields == nil {
		p.UintFields = inputMap
	}
	for k, v := range inputMap {
		p.UintFields[k] = v
	}
}

// AppendUint64Field adds a value to Payload.Uint64Fields; overwrites current value as default behavior.
func (p *Payload) AppendUint64Field(key string, value uint64) {
	if p.Uint64Fields == nil {
		p.Uint64Fields = make(map[string]uint64)
	}
	p.Uint64Fields[key] = value
}

// AppendUint64Fields adds all values of the input map to the payload.UintFields
func (p *Payload) AppendUint64Fields(inputMap map[string]uint64) {
	if inputMap == nil {
		return
	}
	if p.Uint64Fields == nil {
		p.Uint64Fields = inputMap
	}
	for k, v := range inputMap {
		p.Uint64Fields[k] = v
	}
}

// AppendDoubleField adds a value to Payload.DoubleFields; overwrites current value as default behavior.
func (p *Payload) AppendDoubleField(key string, value float64) {
	if p.DoubleFields == nil {
		p.DoubleFields = make(map[string]float64)
	}
	p.DoubleFields[key] = value
}

// AppendDoubleFields adds all values of the input map to the payload.DoubleFields
func (p *Payload) AppendDoubleFields(inputMap map[string]float64) {
	if inputMap == nil {
		return
	}
	if p.DoubleFields == nil {
		p.DoubleFields = inputMap
	}
	for k, v := range inputMap {
		p.DoubleFields[k] = v
	}
}

// AppendFloatField adds a value to Payload.FloatFields; overwrites current value as default behavior.
func (p *Payload) AppendFloatField(key string, value float32) {
	if p.FloatFields == nil {
		p.FloatFields = make(map[string]float32)
	}
	p.FloatFields[key] = value
}

// AppendFloatFields adds all values of the input map to the payload.FloatFields
func (p *Payload) AppendFloatFields(inputMap map[string]float32) {
	if inputMap == nil {
		return
	}
	if p.FloatFields == nil {
		p.FloatFields = inputMap
	}
	for k, v := range inputMap {
		p.FloatFields[k] = v
	}
}

// Proto ; parses payload to protocall buffer
func (p *Payload) Proto() *intrigue.Payload {
	proto := &intrigue.Payload{}
	if p == nil {
		return proto
	}
	if p.Fields != nil {
		m := make(map[string]*intrigue.SubFields)
		for k, v := range p.Fields {
			m[k] = &intrigue.SubFields{
				Sub: v,
			}
		}
		proto.Fields = m
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
	if p.TextFields != nil {
		proto.TextFields = p.TextFields
	}
	if p.BoolFields != nil {
		proto.BoolFields = p.BoolFields
	}
	if p.ByteFields != nil {
		proto.ByteFields = p.ByteFields
	}
	if p.IntFields != nil {
		m := make(map[string]int32)
		for k, v := range p.IntFields {
			m[k] = int32(v)
		}
		proto.IntFields = m
	}
	if p.Int64Fields != nil {
		proto.Int64Fields = p.Int64Fields
	}
	if p.UintFields != nil {
		m := make(map[string]uint32)
		for k, v := range p.IntFields {
			m[k] = uint32(v)
		}
		proto.UintFields = m
	}
	if p.Uint64Fields != nil {
		proto.Uint64Fields = p.Uint64Fields
	}
	if p.DoubleFields != nil {
		proto.DoubleFields = p.DoubleFields
	}
	if p.FloatFields != nil {
		proto.FloatFields = p.FloatFields
	}
	return proto
}

// payloadFromProto ;
func payloadFromProto(proto *intrigue.Payload) *Payload {
	p := &Payload{}
	if proto.Fields != nil {
		m := make(map[string][]string)
		for k, v := range proto.Fields {
			m[k] = v.GetSub()
		}
		p.Fields = m
	}
	if proto.JSON != nil {
		m := make(map[string]interface{})
		for k, v := range p.JSON {
			m[k] = v
		}
		p.JSON = m
	}
	p.TextFields = proto.GetTextFields()
	p.BoolFields = proto.GetBoolFields()
	p.ByteFields = proto.GetByteFields()
	if proto.IntFields != nil {
		m := make(map[string]int)
		for k, v := range proto.IntFields {
			m[k] = int(v)
		}
		p.IntFields = m
	}
	p.Int64Fields = proto.GetInt64Fields()
	if p.UintFields != nil {
		m := make(map[string]uint)
		for k, v := range proto.IntFields {
			m[k] = uint(v)
		}
		p.UintFields = m
	}
	p.Uint64Fields = proto.GetUint64Fields()
	p.DoubleFields = proto.GetDoubleFields()
	p.FloatFields = proto.GetFloatFields()
	return p
}
