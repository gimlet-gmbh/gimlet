package gmbh

import (
	"context"
	"fmt"
	"net"

	"github.com/gimlet-gmbh/gimlet/gproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

/*
 * server.go
 * v 0.0.1
 *
 * Abe Dick
 * November 2018
 */

type _server struct{}

func rpcConnect(address string) {
	list, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	gproto.RegisterCabalServer(s, &_server{})

	reflection.Register(s)
	if err := s.Serve(list); err != nil {
		panic(err)
	}
}

func (s *_server) EphemeralRegisterService(ctx context.Context, in *gproto.RegServReq) (*gproto.RegServRep, error) {
	return &gproto.RegServRep{Status: "invalid operation"}, nil
}

func (s *_server) UnregisterService(ctx context.Context, in *gproto.UnregisterReq) (*gproto.UnregisterResp, error) {
	return &gproto.UnregisterResp{Awk: false}, nil
}

func (s *_server) MakeDataRequest(ctx context.Context, in *gproto.DataReq) (*gproto.DataResp, error) {

	fmt.Println("Recieved data request from: " + in.Req.Sender + " w/ target: " + in.Req.Target)

	responder, err := handleDataRequest(*in.Req)
	if err != nil {
		panic(err)
	}

	reply := &gproto.DataResp{Resp: responder}
	return reply, nil
}

func (s *_server) QueryStatus(ctx context.Context, in *gproto.QueryRequest) (*gproto.QueryResponse, error) {

	response := gproto.QueryResponse{
		Awk:     true,
		Status:  true,
		Details: make(map[string]string),
	}

	return &response, nil
}
