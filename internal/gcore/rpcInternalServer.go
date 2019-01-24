package gcore

import (
	"context"
	"fmt"

	"github.com/rs/xid"

	"github.com/gimlet-gmbh/gimlet/cabal"
	"github.com/gimlet-gmbh/gimlet/notify"
)

/*
 * rpcInternalServer.go
 * Abe Dick
 * Nov 2018
 */

// _server is for gRPC interface fulfilment
type _server struct{}

func (s *_server) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {

	notify.StdMsgLog(fmt.Sprintf("<- Ephemeral Registration Request: %s", in.NewServ.Name))
	service, err := core.serviceHandler.GetService(in.NewServ.Name)
	if err != nil {
		panic(err)
	}

	if service.Static.IsServer {
		notify.StdMsgLog(fmt.Sprintf("-> %s: acknowledged with address: %v", in.NewServ.Name, service.Address))
	} else {
		notify.StdMsgLog(fmt.Sprintf("-> %s: acknowledged", in.NewServ.Name))
	}

	reply := &cabal.RegServRep{
		Status:   "acknowledged",
		ID:       xid.New().String(),
		CorePath: core.ProjectPath,
		Address:  service.Address,
	}

	return reply, nil
}

func (s *_server) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {

	notify.StdMsgLog(fmt.Sprintf("<- Data Request; from: %s; to: %s; method: %s", in.Req.Sender, in.Req.Target, in.Req.Method))

	responder, err := handleDataRequest(*in.Req)
	if err != nil {
		notify.StdMsgLog(fmt.Sprintf("Could not contact: %s", in.Req.Target), 1)
		responder.HadError = true
		responder.ErrorString = "Could not contact target"
	}

	reply := &cabal.DataResp{Resp: responder}
	return reply, nil
}

func (s *_server) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {
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
	address, err := core.serviceHandler.GetAddress(req.Target)
	if err != nil {
		return &cabal.Responder{}, err
	}
	return _makeDataRequest(req, address)
}

func (s *_server) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {
	return nil, nil
}
