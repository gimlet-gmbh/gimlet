package gmbh

/*
 * client.go
 * Abe Dick
 * Nov 2018
 */

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gimlet-gmbh/gimlet/gproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

/**********************************************************************************
** Client
**********************************************************************************/

func getRPCClient() (gproto.CabalClient, error) {
	con, err := grpc.Dial("localhost:59999", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return gproto.NewCabalClient(con), nil

}

func getContextCancel() (context.Context, context.CancelFunc) {
	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return ctx, can
}

func makeRequest() (gproto.CabalClient, context.Context, context.CancelFunc, error) {
	client, err := getRPCClient()
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return client, ctx, can, nil
}

func _ephemeralRegisterService(name string, isClient bool, isServer bool) (string, error) {

	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	request := gproto.RegServReq{
		NewServ: &gproto.NewService{
			Name:     name,
			Aliases:  []string{},
			IsClient: isClient,
			IsServer: isServer,
		},
	}

	reply, err := client.EphemeralRegisterService(ctx, &request)
	if err != nil {
		panic(err)
	}

	fmt.Println(reply.Status + "@" + reply.Address + "@" + reply.CorePath)

	return reply.Address, nil
}

func _makeDataRequest(target string, method string, data string) (Responder, error) {

	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	request := gproto.DataReq{
		Req: &gproto.Request{
			Sender: "test",
			Target: target,
			Method: method,
			Data1:  data,
		},
	}

	reply, err := client.MakeDataRequest(ctx, &request)
	if err != nil {
		// panic(err)
		fmt.Println(fmt.Errorf("%v", err.Error()))

		r := Responder{
			HadError:    true,
			ErrorString: err.Error(),
		}
		return r, err

	}

	return responderFromProto(*reply.Resp), nil
}

func _makeUnregisterRequest(name string) {
	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	_, _ = client.UnregisterService(ctx, &gproto.UnregisterReq{Name: name})
}

/**********************************************************************************
** Server
**********************************************************************************/

// _server implements the coms service using gRPC
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
