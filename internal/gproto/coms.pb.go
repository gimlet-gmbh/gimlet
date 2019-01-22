// Code generated by protoc-gen-go. DO NOT EDIT.
// source: coms.proto

package gproto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type QueryRequest_QueryLevel int32

const (
	// Alive Ping
	QueryRequest_STATUS QueryRequest_QueryLevel = 0
	// Report any errors w/ messages
	QueryRequest_REPORT QueryRequest_QueryLevel = 1
	// Report start time, errors, etc.
	QueryRequest_FULL QueryRequest_QueryLevel = 2
)

var QueryRequest_QueryLevel_name = map[int32]string{
	0: "STATUS",
	1: "REPORT",
	2: "FULL",
}

var QueryRequest_QueryLevel_value = map[string]int32{
	"STATUS": 0,
	"REPORT": 1,
	"FULL":   2,
}

func (x QueryRequest_QueryLevel) String() string {
	return proto.EnumName(QueryRequest_QueryLevel_name, int32(x))
}

func (QueryRequest_QueryLevel) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{10, 0}
}

type RegServReq struct {
	// The service to register.
	NewServ              *NewService `protobuf:"bytes,1,opt,name=NewServ,proto3" json:"NewServ,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *RegServReq) Reset()         { *m = RegServReq{} }
func (m *RegServReq) String() string { return proto.CompactTextString(m) }
func (*RegServReq) ProtoMessage()    {}
func (*RegServReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{0}
}

func (m *RegServReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegServReq.Unmarshal(m, b)
}
func (m *RegServReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegServReq.Marshal(b, m, deterministic)
}
func (m *RegServReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegServReq.Merge(m, src)
}
func (m *RegServReq) XXX_Size() int {
	return xxx_messageInfo_RegServReq.Size(m)
}
func (m *RegServReq) XXX_DiscardUnknown() {
	xxx_messageInfo_RegServReq.DiscardUnknown(m)
}

var xxx_messageInfo_RegServReq proto.InternalMessageInfo

func (m *RegServReq) GetNewServ() *NewService {
	if m != nil {
		return m.NewServ
	}
	return nil
}

type RegServRep struct {
	// The status of the server
	Status string `protobuf:"bytes,1,opt,name=Status,proto3" json:"Status,omitempty"`
	// The address to start new service host on
	Address string `protobuf:"bytes,3,opt,name=Address,proto3" json:"Address,omitempty"`
	// The id that CORE assigned to this service
	ID string `protobuf:"bytes,4,opt,name=ID,proto3" json:"ID,omitempty"`
	// The path to where the core instance is running
	// That is, the where the project root dir is.
	CorePath string `protobuf:"bytes,5,opt,name=CorePath,proto3" json:"CorePath,omitempty"`
	// An array of all services.
	// *NOTE* This will need improvement as a full walk of all
	//        discoverable services must be done by walking each
	//        indice of the arrach and then walk the "Aliases"
	//        field for a match.
	ServiceTable         []*S2S   `protobuf:"bytes,2,rep,name=ServiceTable,proto3" json:"ServiceTable,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegServRep) Reset()         { *m = RegServRep{} }
func (m *RegServRep) String() string { return proto.CompactTextString(m) }
func (*RegServRep) ProtoMessage()    {}
func (*RegServRep) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{1}
}

func (m *RegServRep) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegServRep.Unmarshal(m, b)
}
func (m *RegServRep) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegServRep.Marshal(b, m, deterministic)
}
func (m *RegServRep) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegServRep.Merge(m, src)
}
func (m *RegServRep) XXX_Size() int {
	return xxx_messageInfo_RegServRep.Size(m)
}
func (m *RegServRep) XXX_DiscardUnknown() {
	xxx_messageInfo_RegServRep.DiscardUnknown(m)
}

var xxx_messageInfo_RegServRep proto.InternalMessageInfo

func (m *RegServRep) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *RegServRep) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *RegServRep) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *RegServRep) GetCorePath() string {
	if m != nil {
		return m.CorePath
	}
	return ""
}

func (m *RegServRep) GetServiceTable() []*S2S {
	if m != nil {
		return m.ServiceTable
	}
	return nil
}

//
// NewService {message}
//
// The proto representation of a new service
type NewService struct {
	// The name of the service
	Name string `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	// Any aliases to recognize the service
	Aliases []string `protobuf:"bytes,2,rep,name=Aliases,proto3" json:"Aliases,omitempty"`
	// True if service can make data requests
	IsClient bool `protobuf:"varint,3,opt,name=IsClient,proto3" json:"IsClient,omitempty"`
	// True if service can handle data requests
	IsServer             bool     `protobuf:"varint,4,opt,name=IsServer,proto3" json:"IsServer,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NewService) Reset()         { *m = NewService{} }
func (m *NewService) String() string { return proto.CompactTextString(m) }
func (*NewService) ProtoMessage()    {}
func (*NewService) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{2}
}

func (m *NewService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NewService.Unmarshal(m, b)
}
func (m *NewService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NewService.Marshal(b, m, deterministic)
}
func (m *NewService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NewService.Merge(m, src)
}
func (m *NewService) XXX_Size() int {
	return xxx_messageInfo_NewService.Size(m)
}
func (m *NewService) XXX_DiscardUnknown() {
	xxx_messageInfo_NewService.DiscardUnknown(m)
}

var xxx_messageInfo_NewService proto.InternalMessageInfo

func (m *NewService) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *NewService) GetAliases() []string {
	if m != nil {
		return m.Aliases
	}
	return nil
}

func (m *NewService) GetIsClient() bool {
	if m != nil {
		return m.IsClient
	}
	return false
}

func (m *NewService) GetIsServer() bool {
	if m != nil {
		return m.IsServer
	}
	return false
}

//
// S2S - Service To Service
//
// The proto representation of a service to be used for inter-service
// communications that bypass CabalCORE
type S2S struct {
	// The name of the service
	Name string `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	// Any aliases to recognize the service
	Aliases []string `protobuf:"bytes,2,rep,name=Aliases,proto3" json:"Aliases,omitempty"`
	// The complete location of the service.
	// Should be expressed as a complete URL if using TCP. If using load-balancers
	// it should be the location of the load balancer.
	Location             string   `protobuf:"bytes,3,opt,name=Location,proto3" json:"Location,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *S2S) Reset()         { *m = S2S{} }
func (m *S2S) String() string { return proto.CompactTextString(m) }
func (*S2S) ProtoMessage()    {}
func (*S2S) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{3}
}

func (m *S2S) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_S2S.Unmarshal(m, b)
}
func (m *S2S) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_S2S.Marshal(b, m, deterministic)
}
func (m *S2S) XXX_Merge(src proto.Message) {
	xxx_messageInfo_S2S.Merge(m, src)
}
func (m *S2S) XXX_Size() int {
	return xxx_messageInfo_S2S.Size(m)
}
func (m *S2S) XXX_DiscardUnknown() {
	xxx_messageInfo_S2S.DiscardUnknown(m)
}

var xxx_messageInfo_S2S proto.InternalMessageInfo

func (m *S2S) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *S2S) GetAliases() []string {
	if m != nil {
		return m.Aliases
	}
	return nil
}

func (m *S2S) GetLocation() string {
	if m != nil {
		return m.Location
	}
	return ""
}

type DataReq struct {
	// The Cabal Request object
	Req                  *Request `protobuf:"bytes,1,opt,name=Req,proto3" json:"Req,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DataReq) Reset()         { *m = DataReq{} }
func (m *DataReq) String() string { return proto.CompactTextString(m) }
func (*DataReq) ProtoMessage()    {}
func (*DataReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{4}
}

func (m *DataReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataReq.Unmarshal(m, b)
}
func (m *DataReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataReq.Marshal(b, m, deterministic)
}
func (m *DataReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataReq.Merge(m, src)
}
func (m *DataReq) XXX_Size() int {
	return xxx_messageInfo_DataReq.Size(m)
}
func (m *DataReq) XXX_DiscardUnknown() {
	xxx_messageInfo_DataReq.DiscardUnknown(m)
}

var xxx_messageInfo_DataReq proto.InternalMessageInfo

func (m *DataReq) GetReq() *Request {
	if m != nil {
		return m.Req
	}
	return nil
}

type DataResp struct {
	// The Cabal Response object
	Resp                 *Responder `protobuf:"bytes,1,opt,name=Resp,proto3" json:"Resp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *DataResp) Reset()         { *m = DataResp{} }
func (m *DataResp) String() string { return proto.CompactTextString(m) }
func (*DataResp) ProtoMessage()    {}
func (*DataResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{5}
}

func (m *DataResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataResp.Unmarshal(m, b)
}
func (m *DataResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataResp.Marshal(b, m, deterministic)
}
func (m *DataResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataResp.Merge(m, src)
}
func (m *DataResp) XXX_Size() int {
	return xxx_messageInfo_DataResp.Size(m)
}
func (m *DataResp) XXX_DiscardUnknown() {
	xxx_messageInfo_DataResp.DiscardUnknown(m)
}

var xxx_messageInfo_DataResp proto.InternalMessageInfo

func (m *DataResp) GetResp() *Responder {
	if m != nil {
		return m.Resp
	}
	return nil
}

type Request struct {
	// Name of the sender
	Sender string `protobuf:"bytes,1,opt,name=Sender,proto3" json:"Sender,omitempty"`
	// Name or alias of the target
	Target string `protobuf:"bytes,2,opt,name=Target,proto3" json:"Target,omitempty"`
	// The handler to invoke on target entry
	Method string `protobuf:"bytes,3,opt,name=Method,proto3" json:"Method,omitempty"`
	// The data to send
	// *NOTE* This will be converted to a much more sophisticated method later.
	Data1                string   `protobuf:"bytes,50,opt,name=Data1,proto3" json:"Data1,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{6}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *Request) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func (m *Request) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *Request) GetData1() string {
	if m != nil {
		return m.Data1
	}
	return ""
}

type Responder struct {
	// The result from target
	// *NOTE* This will be converted to a much more sophisticated method later.
	Result string `protobuf:"bytes,50,opt,name=Result,proto3" json:"Result,omitempty"`
	// If HadError, this will contain the error message associated
	ErrorString string `protobuf:"bytes,98,opt,name=ErrorString,proto3" json:"ErrorString,omitempty"`
	// True if there was an error that did NOT involve tcp
	HadError             bool     `protobuf:"varint,99,opt,name=HadError,proto3" json:"HadError,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Responder) Reset()         { *m = Responder{} }
func (m *Responder) String() string { return proto.CompactTextString(m) }
func (*Responder) ProtoMessage()    {}
func (*Responder) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{7}
}

func (m *Responder) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Responder.Unmarshal(m, b)
}
func (m *Responder) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Responder.Marshal(b, m, deterministic)
}
func (m *Responder) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Responder.Merge(m, src)
}
func (m *Responder) XXX_Size() int {
	return xxx_messageInfo_Responder.Size(m)
}
func (m *Responder) XXX_DiscardUnknown() {
	xxx_messageInfo_Responder.DiscardUnknown(m)
}

var xxx_messageInfo_Responder proto.InternalMessageInfo

func (m *Responder) GetResult() string {
	if m != nil {
		return m.Result
	}
	return ""
}

func (m *Responder) GetErrorString() string {
	if m != nil {
		return m.ErrorString
	}
	return ""
}

func (m *Responder) GetHadError() bool {
	if m != nil {
		return m.HadError
	}
	return false
}

type UnregisterReq struct {
	// The id of the service to unregister
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The name of the service to unregister
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// the address of the service to unregister
	Address              string   `protobuf:"bytes,3,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UnregisterReq) Reset()         { *m = UnregisterReq{} }
func (m *UnregisterReq) String() string { return proto.CompactTextString(m) }
func (*UnregisterReq) ProtoMessage()    {}
func (*UnregisterReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{8}
}

func (m *UnregisterReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UnregisterReq.Unmarshal(m, b)
}
func (m *UnregisterReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UnregisterReq.Marshal(b, m, deterministic)
}
func (m *UnregisterReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnregisterReq.Merge(m, src)
}
func (m *UnregisterReq) XXX_Size() int {
	return xxx_messageInfo_UnregisterReq.Size(m)
}
func (m *UnregisterReq) XXX_DiscardUnknown() {
	xxx_messageInfo_UnregisterReq.DiscardUnknown(m)
}

var xxx_messageInfo_UnregisterReq proto.InternalMessageInfo

func (m *UnregisterReq) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *UnregisterReq) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UnregisterReq) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type UnregisterResp struct {
	// If CORE awknowledged the unregister reuqest
	Awk                  bool     `protobuf:"varint,1,opt,name=awk,proto3" json:"awk,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UnregisterResp) Reset()         { *m = UnregisterResp{} }
func (m *UnregisterResp) String() string { return proto.CompactTextString(m) }
func (*UnregisterResp) ProtoMessage()    {}
func (*UnregisterResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{9}
}

func (m *UnregisterResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UnregisterResp.Unmarshal(m, b)
}
func (m *UnregisterResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UnregisterResp.Marshal(b, m, deterministic)
}
func (m *UnregisterResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnregisterResp.Merge(m, src)
}
func (m *UnregisterResp) XXX_Size() int {
	return xxx_messageInfo_UnregisterResp.Size(m)
}
func (m *UnregisterResp) XXX_DiscardUnknown() {
	xxx_messageInfo_UnregisterResp.DiscardUnknown(m)
}

var xxx_messageInfo_UnregisterResp proto.InternalMessageInfo

func (m *UnregisterResp) GetAwk() bool {
	if m != nil {
		return m.Awk
	}
	return false
}

type QueryRequest struct {
	// The type of query -- see the enumerated type above
	Query                QueryRequest_QueryLevel `protobuf:"varint,1,opt,name=Query,proto3,enum=gproto.QueryRequest_QueryLevel" json:"Query,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *QueryRequest) Reset()         { *m = QueryRequest{} }
func (m *QueryRequest) String() string { return proto.CompactTextString(m) }
func (*QueryRequest) ProtoMessage()    {}
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{10}
}

func (m *QueryRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryRequest.Unmarshal(m, b)
}
func (m *QueryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryRequest.Marshal(b, m, deterministic)
}
func (m *QueryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryRequest.Merge(m, src)
}
func (m *QueryRequest) XXX_Size() int {
	return xxx_messageInfo_QueryRequest.Size(m)
}
func (m *QueryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryRequest proto.InternalMessageInfo

func (m *QueryRequest) GetQuery() QueryRequest_QueryLevel {
	if m != nil {
		return m.Query
	}
	return QueryRequest_STATUS
}

type QueryResponse struct {
	// if the process awknoledges request
	Awk bool `protobuf:"varint,1,opt,name=Awk,proto3" json:"Awk,omitempty"`
	// True if running without errors
	Status bool `protobuf:"varint,2,opt,name=Status,proto3" json:"Status,omitempty"`
	// The details map to pass
	Details map[string]string `protobuf:"bytes,3,rep,name=Details,proto3" json:"Details,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The error messages reported from gmbh gplugin
	Errors               []string `protobuf:"bytes,4,rep,name=Errors,proto3" json:"Errors,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryResponse) Reset()         { *m = QueryResponse{} }
func (m *QueryResponse) String() string { return proto.CompactTextString(m) }
func (*QueryResponse) ProtoMessage()    {}
func (*QueryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_49604cccc55913fc, []int{11}
}

func (m *QueryResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryResponse.Unmarshal(m, b)
}
func (m *QueryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryResponse.Marshal(b, m, deterministic)
}
func (m *QueryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryResponse.Merge(m, src)
}
func (m *QueryResponse) XXX_Size() int {
	return xxx_messageInfo_QueryResponse.Size(m)
}
func (m *QueryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryResponse proto.InternalMessageInfo

func (m *QueryResponse) GetAwk() bool {
	if m != nil {
		return m.Awk
	}
	return false
}

func (m *QueryResponse) GetStatus() bool {
	if m != nil {
		return m.Status
	}
	return false
}

func (m *QueryResponse) GetDetails() map[string]string {
	if m != nil {
		return m.Details
	}
	return nil
}

func (m *QueryResponse) GetErrors() []string {
	if m != nil {
		return m.Errors
	}
	return nil
}

func init() {
	proto.RegisterEnum("gproto.QueryRequest_QueryLevel", QueryRequest_QueryLevel_name, QueryRequest_QueryLevel_value)
	proto.RegisterType((*RegServReq)(nil), "gproto.RegServReq")
	proto.RegisterType((*RegServRep)(nil), "gproto.RegServRep")
	proto.RegisterType((*NewService)(nil), "gproto.NewService")
	proto.RegisterType((*S2S)(nil), "gproto.S2S")
	proto.RegisterType((*DataReq)(nil), "gproto.DataReq")
	proto.RegisterType((*DataResp)(nil), "gproto.DataResp")
	proto.RegisterType((*Request)(nil), "gproto.Request")
	proto.RegisterType((*Responder)(nil), "gproto.Responder")
	proto.RegisterType((*UnregisterReq)(nil), "gproto.UnregisterReq")
	proto.RegisterType((*UnregisterResp)(nil), "gproto.UnregisterResp")
	proto.RegisterType((*QueryRequest)(nil), "gproto.QueryRequest")
	proto.RegisterType((*QueryResponse)(nil), "gproto.QueryResponse")
	proto.RegisterMapType((map[string]string)(nil), "gproto.QueryResponse.DetailsEntry")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// CabalClient is the client API for Cabal service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CabalClient interface {
	//
	// EphemeralRegisterService
	//
	// This should be used to for attaching services that use the generic data scheme and are considered
	// to be ephemeral to CabalCORE
	EphemeralRegisterService(ctx context.Context, in *RegServReq, opts ...grpc.CallOption) (*RegServRep, error)
	//
	// MakeDataRequest
	//
	// Make all data requests between services through this method.
	// *NOTE* This is a daft spec.
	MakeDataRequest(ctx context.Context, in *DataReq, opts ...grpc.CallOption) (*DataResp, error)
	//
	// UnregisterService
	//
	// This should be used for graceful shutdown notice to CORE.
	UnregisterService(ctx context.Context, in *UnregisterReq, opts ...grpc.CallOption) (*UnregisterResp, error)
	//
	// QueryStatus
	//
	// Use this to query the status of all processes attached to CORE.
	QueryStatus(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error)
}

type cabalClient struct {
	cc *grpc.ClientConn
}

func NewCabalClient(cc *grpc.ClientConn) CabalClient {
	return &cabalClient{cc}
}

func (c *cabalClient) EphemeralRegisterService(ctx context.Context, in *RegServReq, opts ...grpc.CallOption) (*RegServRep, error) {
	out := new(RegServRep)
	err := c.cc.Invoke(ctx, "/gproto.Cabal/EphemeralRegisterService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cabalClient) MakeDataRequest(ctx context.Context, in *DataReq, opts ...grpc.CallOption) (*DataResp, error) {
	out := new(DataResp)
	err := c.cc.Invoke(ctx, "/gproto.Cabal/MakeDataRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cabalClient) UnregisterService(ctx context.Context, in *UnregisterReq, opts ...grpc.CallOption) (*UnregisterResp, error) {
	out := new(UnregisterResp)
	err := c.cc.Invoke(ctx, "/gproto.Cabal/UnregisterService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cabalClient) QueryStatus(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error) {
	out := new(QueryResponse)
	err := c.cc.Invoke(ctx, "/gproto.Cabal/QueryStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CabalServer is the server API for Cabal service.
type CabalServer interface {
	//
	// EphemeralRegisterService
	//
	// This should be used to for attaching services that use the generic data scheme and are considered
	// to be ephemeral to CabalCORE
	EphemeralRegisterService(context.Context, *RegServReq) (*RegServRep, error)
	//
	// MakeDataRequest
	//
	// Make all data requests between services through this method.
	// *NOTE* This is a daft spec.
	MakeDataRequest(context.Context, *DataReq) (*DataResp, error)
	//
	// UnregisterService
	//
	// This should be used for graceful shutdown notice to CORE.
	UnregisterService(context.Context, *UnregisterReq) (*UnregisterResp, error)
	//
	// QueryStatus
	//
	// Use this to query the status of all processes attached to CORE.
	QueryStatus(context.Context, *QueryRequest) (*QueryResponse, error)
}

func RegisterCabalServer(s *grpc.Server, srv CabalServer) {
	s.RegisterService(&_Cabal_serviceDesc, srv)
}

func _Cabal_EphemeralRegisterService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegServReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CabalServer).EphemeralRegisterService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gproto.Cabal/EphemeralRegisterService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CabalServer).EphemeralRegisterService(ctx, req.(*RegServReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cabal_MakeDataRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DataReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CabalServer).MakeDataRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gproto.Cabal/MakeDataRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CabalServer).MakeDataRequest(ctx, req.(*DataReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cabal_UnregisterService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnregisterReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CabalServer).UnregisterService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gproto.Cabal/UnregisterService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CabalServer).UnregisterService(ctx, req.(*UnregisterReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cabal_QueryStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CabalServer).QueryStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gproto.Cabal/QueryStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CabalServer).QueryStatus(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Cabal_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gproto.Cabal",
	HandlerType: (*CabalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EphemeralRegisterService",
			Handler:    _Cabal_EphemeralRegisterService_Handler,
		},
		{
			MethodName: "MakeDataRequest",
			Handler:    _Cabal_MakeDataRequest_Handler,
		},
		{
			MethodName: "UnregisterService",
			Handler:    _Cabal_UnregisterService_Handler,
		},
		{
			MethodName: "QueryStatus",
			Handler:    _Cabal_QueryStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coms.proto",
}

func init() { proto.RegisterFile("coms.proto", fileDescriptor_49604cccc55913fc) }

var fileDescriptor_49604cccc55913fc = []byte{
	// 683 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xdb, 0x4e, 0x1b, 0x31,
	0x10, 0x25, 0x9b, 0x2b, 0x13, 0x08, 0xc1, 0x02, 0xb4, 0xca, 0x4b, 0xa9, 0xa5, 0x4a, 0x3c, 0xa0,
	0x54, 0xa4, 0x6a, 0x55, 0x21, 0x5e, 0x22, 0x92, 0xaa, 0x48, 0xe1, 0x52, 0x6f, 0xf8, 0x00, 0x27,
	0x19, 0x85, 0x55, 0x96, 0x6c, 0xb0, 0x1d, 0x10, 0x9f, 0xd0, 0x6f, 0xe8, 0x27, 0xf5, 0xa7, 0x2a,
	0x8f, 0xbd, 0x9b, 0xa4, 0xe5, 0xa5, 0x4f, 0xf1, 0x39, 0x73, 0x3c, 0x7b, 0x7c, 0xc6, 0x0e, 0xc0,
	0x38, 0x7d, 0xd4, 0xed, 0x85, 0x4a, 0x4d, 0xca, 0x2a, 0x53, 0xfa, 0xe5, 0xe7, 0x00, 0x02, 0xa7,
	0x11, 0xaa, 0x67, 0x81, 0x4f, 0xec, 0x14, 0xaa, 0x37, 0xf8, 0x62, 0x51, 0x58, 0x38, 0x2e, 0x9c,
	0xd4, 0x3b, 0xac, 0xed, 0x74, 0x6d, 0x4f, 0xc7, 0x63, 0x14, 0x99, 0x84, 0xff, 0x2a, 0xac, 0x6d,
	0x5e, 0xb0, 0x23, 0xa8, 0x44, 0x46, 0x9a, 0xa5, 0xa6, 0xbd, 0xdb, 0xc2, 0x23, 0x16, 0x42, 0xb5,
	0x3b, 0x99, 0x28, 0xd4, 0x3a, 0x2c, 0x52, 0x21, 0x83, 0xac, 0x01, 0xc1, 0x55, 0x2f, 0x2c, 0x11,
	0x19, 0x5c, 0xf5, 0x58, 0x0b, 0x6a, 0x97, 0xa9, 0xc2, 0x3b, 0x69, 0x1e, 0xc2, 0x32, 0xb1, 0x39,
	0x66, 0x1f, 0x61, 0xc7, 0x1b, 0x18, 0xca, 0x51, 0x82, 0x61, 0x70, 0x5c, 0x3c, 0xa9, 0x77, 0xea,
	0x99, 0xbf, 0xa8, 0x13, 0x89, 0x0d, 0x01, 0x57, 0x00, 0x2b, 0xd3, 0x8c, 0x41, 0xe9, 0x46, 0x3e,
	0xa2, 0xb7, 0x46, 0x6b, 0x32, 0x96, 0xc4, 0x52, 0xa3, 0xa6, 0x6e, 0xd6, 0x98, 0x83, 0xd6, 0xc8,
	0x95, 0xbe, 0x4c, 0x62, 0x9c, 0x1b, 0xf2, 0x5c, 0x13, 0x39, 0x76, 0x35, 0xdb, 0x16, 0x15, 0x59,
	0xa7, 0x9a, 0xc3, 0xfc, 0x16, 0x8a, 0x51, 0x27, 0xfa, 0xff, 0x8f, 0x0d, 0xd2, 0xb1, 0x34, 0x71,
	0x3a, 0xf7, 0x01, 0xe5, 0x98, 0x9f, 0x42, 0xb5, 0x27, 0x8d, 0xb4, 0xb3, 0x79, 0x0f, 0x45, 0x81,
	0x4f, 0x7e, 0x2e, 0x7b, 0xd9, 0xb9, 0x05, 0x3e, 0x2d, 0x51, 0x1b, 0x61, 0x6b, 0xfc, 0x0c, 0x6a,
	0x4e, 0xad, 0x17, 0xec, 0x03, 0x94, 0xec, 0xaf, 0xd7, 0xef, 0xaf, 0xf4, 0x7a, 0x91, 0xce, 0x27,
	0xa8, 0x04, 0x95, 0xf9, 0x14, 0xaa, 0xbe, 0x05, 0xcd, 0x0f, 0x6d, 0x29, 0x9f, 0x1f, 0x21, 0xcb,
	0x0f, 0xa5, 0x9a, 0xa2, 0x09, 0x03, 0xc7, 0x3b, 0x64, 0xf9, 0x6b, 0x34, 0x0f, 0xe9, 0xc4, 0xbb,
	0xf6, 0x88, 0x1d, 0x40, 0xd9, 0xba, 0x38, 0x0b, 0x3b, 0x44, 0x3b, 0xc0, 0x25, 0x6c, 0xe7, 0xdf,
	0xb6, 0x5b, 0x05, 0xea, 0x65, 0x62, 0xbc, 0xc6, 0x23, 0x76, 0x0c, 0xf5, 0xbe, 0x52, 0xa9, 0x8a,
	0x8c, 0x8a, 0xe7, 0xd3, 0x70, 0x44, 0xc5, 0x75, 0xca, 0x86, 0xf5, 0x5d, 0x4e, 0x88, 0x09, 0xc7,
	0x2e, 0xfd, 0x0c, 0xf3, 0x6b, 0xd8, 0xbd, 0x9f, 0x2b, 0x9c, 0xc6, 0xda, 0xa0, 0xb2, 0x91, 0x35,
	0x20, 0x88, 0x27, 0xfe, 0x34, 0x41, 0x3c, 0xb1, 0x73, 0x99, 0xdb, 0xb9, 0xb8, 0x73, 0xd0, 0xda,
	0xce, 0x45, 0x6e, 0xde, 0x4e, 0x0f, 0x39, 0x87, 0xc6, 0x7a, 0x3b, 0xbd, 0x60, 0x4d, 0x28, 0xca,
	0x97, 0x19, 0x35, 0xac, 0x09, 0xbb, 0xe4, 0x4b, 0xd8, 0xf9, 0xb1, 0x44, 0xf5, 0x9a, 0x65, 0xf8,
	0x19, 0xca, 0x84, 0x49, 0xd3, 0xe8, 0xbc, 0xcb, 0x62, 0x5f, 0x17, 0x39, 0x30, 0xc0, 0x67, 0x4c,
	0x84, 0x53, 0xf3, 0x36, 0xc0, 0x8a, 0x64, 0x00, 0x95, 0x68, 0xd8, 0x1d, 0xde, 0x47, 0xcd, 0x2d,
	0xbb, 0x16, 0xfd, 0xbb, 0x5b, 0x31, 0x6c, 0x16, 0x58, 0x0d, 0x4a, 0xdf, 0xee, 0x07, 0x83, 0x66,
	0xc0, 0x7f, 0x17, 0x60, 0xd7, 0xb7, 0xb4, 0x91, 0x6a, 0xb4, 0xd6, 0xba, 0x2b, 0x6b, 0xdd, 0x97,
	0xd9, 0xda, 0x73, 0x0c, 0x88, 0xcc, 0x9e, 0xe3, 0x05, 0x54, 0x7b, 0x68, 0x64, 0x9c, 0xd8, 0x03,
	0xdb, 0x37, 0xc4, 0xff, 0x32, 0xe9, 0x3a, 0xb6, 0xbd, 0xa8, 0x3f, 0x37, 0xea, 0x55, 0x64, 0x5b,
	0x6c, 0x57, 0x0a, 0x5b, 0x87, 0x25, 0xba, 0xc5, 0x1e, 0xb5, 0xce, 0x61, 0x67, 0x7d, 0x83, 0xf5,
	0x33, 0xc3, 0x57, 0x9f, 0xbd, 0x5d, 0xda, 0x6b, 0xf1, 0x2c, 0x93, 0x65, 0x96, 0xbe, 0x03, 0xe7,
	0xc1, 0xd7, 0x42, 0xe7, 0x67, 0x00, 0xe5, 0x4b, 0x39, 0x92, 0x09, 0xeb, 0x41, 0xd8, 0x5f, 0x3c,
	0xe0, 0x23, 0x2a, 0x99, 0x08, 0x9f, 0x7c, 0xfe, 0x82, 0x57, 0x57, 0x38, 0xfb, 0xbf, 0x6a, 0xfd,
	0xcb, 0x2d, 0xf8, 0x16, 0xfb, 0x02, 0x7b, 0xd7, 0x72, 0x86, 0xfe, 0xe1, 0xd0, 0x5c, 0xf2, 0xf7,
	0xe2, 0xc9, 0x56, 0x73, 0x93, 0xd0, 0x76, 0x5f, 0x0f, 0xf6, 0x57, 0x03, 0xcf, 0x3e, 0x7b, 0x98,
	0x09, 0x37, 0xae, 0x56, 0xeb, 0xe8, 0x2d, 0x9a, 0xba, 0x5c, 0x40, 0x9d, 0x82, 0xf4, 0x71, 0x1f,
	0xbc, 0x75, 0x05, 0x5a, 0x87, 0x6f, 0x66, 0xce, 0xb7, 0x46, 0x15, 0xa2, 0x3f, 0xfd, 0x09, 0x00,
	0x00, 0xff, 0xff, 0x00, 0x74, 0xe9, 0x10, 0xac, 0x05, 0x00, 0x00,
}
