package main

/*
 * cabal.go
 * Implements the gRPC server and client for the gmbhCore Cabal Server
 * Abe Dick
 * Nov 2018
 */

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/rs/xid"
	"google.golang.org/grpc"
)

/////////////////////////////////////////////////////////////////////////
// CLIENT
/////////////////////////////////////////////////////////////////////////

func getRPCClient(address string) (cabal.CabalClient, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return cabal.NewCabalClient(con), nil
}

func getContextCancel() (context.Context, context.CancelFunc) {
	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return ctx, can
}

func makeRequest(address string) (cabal.CabalClient, context.Context, context.CancelFunc, error) {
	client, err := getRPCClient(address)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return client, ctx, can, nil
}

func _makeDataRequest(req cabal.Request, address string) (*cabal.Responder, error) {
	client, ctx, can, err := makeRequest(address)
	if err != nil {
		panic(err)
	}
	defer can()

	request := cabal.DataReq{
		Req: &req,
	}

	reply, err := client.MakeDataRequest(ctx, &request)
	if err != nil {
		notify.StdMsgErr(err.Error(), 1)
		r := cabal.Responder{
			HadError:    true,
			ErrorString: err.Error(),
		}
		return &r, nil
	}

	return reply.Resp, nil
}

func requestQueryData(address string) (*cabal.QueryResponse, error) {
	client, ctx, can, err := makeRequest(address)
	if err != nil {
		panic(err)
	}
	defer can()

	req := cabal.QueryRequest{
		Query: cabal.QueryRequest_STATUS,
	}

	// reply, err := client.QueryStatus(ctx, &req)
	return client.QueryStatus(ctx, &req)
}

/////////////////////////////////////////////////////////////////////////
// SERVER
/////////////////////////////////////////////////////////////////////////

// cabalServer is for gRPC interface for the gmbhCore service coms server
type cabalServer struct{}

func (s *cabalServer) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {

	service, err := core.Router.LookupService(in.NewServ.GetName())
	if err != nil {
		if err.Error() == "router.LookupService.nameNotFound" {
			panic(err)
		}
	}

	if !core.Config.Daemon && core.Config.Verbose {
		notify.StdMsgLog(fmt.Sprintf("<- Ephemeral Registration Request: %s", in.NewServ.GetName()))
		if service.Static.IsServer {
			notify.StdMsgLog(fmt.Sprintf("-> %s: acknowledged with address: %v", in.NewServ.GetName(), service.Address))
		} else {
			notify.StdMsgLog(fmt.Sprintf("-> %s: acknowledged", in.NewServ.GetName()))
		}
	}

	reply := &cabal.RegServRep{
		Status:   "acknowledged",
		ID:       xid.New().String(),
		CorePath: core.ProjectPath,
		Address:  service.Address,
	}

	return reply, nil
}

func (s *cabalServer) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {
	cc, err := getCore()
	if err != nil {
		return nil, errors.New("gmbh system error, could not locate instance of core")
	}
	if !cc.Config.Daemon {
		notify.StdMsgLog(fmt.Sprintf("<- Data Request; from: %s; to: %s; method: %s", in.Req.GetSender(), in.Req.GetTarget(), in.Req.GetMethod()))

	}
	responder, err := handleDataRequest(*in.Req)
	if err != nil {
		if !cc.Config.Daemon {
			notify.StdMsgLog(fmt.Sprintf("Could not contact: %s", in.Req.Target), 1)
		}
		responder.HadError = true
		responder.ErrorString = "Could not contact target"
	}

	reply := &cabal.DataResp{Resp: responder}
	return reply, nil
}

func (s *cabalServer) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {
	// printDebug("Received unregister request")
	// printDebug("\tName: " + in.Name)

	reply := &cabal.UnregisterResp{Awk: false}
	// err := removeServiceFromList(in.Name, -1)
	// if err != nil {
	// reply.Awk = false
	// }
	// listAllServices()
	return reply, nil
}

func handleDataRequest(req cabal.Request) (*cabal.Responder, error) {
	address, err := core.Router.LookupAddress(req.GetTarget())
	if err != nil {
		return &cabal.Responder{}, err
	}
	if !core.Config.Daemon && core.Config.Verbose {
		notify.StdMsgLog(fmt.Sprintf("   Target address on file: %v", address))
	}
	return _makeDataRequest(req, address)
}

func (s *cabalServer) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {
	return nil, nil
}
