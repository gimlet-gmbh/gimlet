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
	"strconv"
	"time"

	"github.com/gimlet-gmbh/gimlet/cabal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

/**********************************************************************************
** Client
**********************************************************************************/

func getRPCClient() (cabal.CabalClient, error) {
	con, err := grpc.Dial("localhost:59999", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return cabal.NewCabalClient(con), nil

}

func getContextCancel() (context.Context, context.CancelFunc) {
	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return ctx, can
}

func makeRequest() (cabal.CabalClient, context.Context, context.CancelFunc, error) {
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

	request := cabal.RegServReq{
		NewServ: &cabal.NewService{
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

	request := cabal.DataReq{
		Req: &cabal.Request{
			Sender: "test",
			Target: target,
			Method: method,
			Data1:  data,
		},
	}

	mcs := strconv.Itoa(g.msgCounter)
	g.msgCounter++
	dlog("<==" + mcs + "== target: " + target + ", method: " + method)

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
	dlog(" ==" + mcs + "==> result: " + reply.Resp.Result + ", errors?: " + reply.Resp.ErrorString)

	return responderFromProto(*reply.Resp), nil
}

func _makeUnregisterRequest(name string) {
	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	_, _ = client.UnregisterService(ctx, &cabal.UnregisterReq{Name: name})
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
	cabal.RegisterCabalServer(s, &_server{})

	reflection.Register(s)
	if err := s.Serve(list); err != nil {
		panic(err)
	}
}

func (s *_server) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {
	return &cabal.RegServRep{Status: "invalid operation"}, nil
}

func (s *_server) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {
	return &cabal.UnregisterResp{Awk: false}, nil
}

func (s *_server) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {

	mcs := strconv.Itoa(g.msgCounter)
	g.msgCounter++
	dlog("==" + mcs + "==> from: " + in.Req.Sender + ", method: " + in.Req.Method)

	responder, err := handleDataRequest(*in.Req)
	if err != nil {
		panic(err)
	}

	reply := &cabal.DataResp{Resp: responder}
	return reply, nil
}

func (s *_server) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {

	response := cabal.QueryResponse{
		Awk:     true,
		Status:  true,
		Details: make(map[string]string),
	}

	return &response, nil
}
