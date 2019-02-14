package main

import (
	"context"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
)

func v(msg string) {
	notify.StdMsgBlueNoPrompt(" [cbl] " + msg)

}

// cabalServer is for gRPC interface for the gmbhCore service coms server
type cabalServer struct{}

func (s *cabalServer) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {

	rv("-> Incoming Registration;")
	rv("   Name=%s; Aliases=%s;", in.GetNewServ().GetName(), in.GetNewServ().GetAliases())

	c, err := GetCore()
	if err != nil {
		return &cabal.RegServRep{Status: "internal error"}, nil
	}

	ns, err := c.Router.AddService(in.GetNewServ().GetName(), in.GetNewServ().GetAliases())
	if err != nil {
		return &cabal.RegServRep{Status: "error=" + err.Error()}, nil
	}

	reply := &cabal.RegServRep{
		Status: "acknowledged",

		Address:  ns.Address,
		ID:       ns.ID,
		CorePath: c.ProjectPath,
	}
	return reply, nil

}

func (s *cabalServer) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {
	return &cabal.DataResp{}, nil
	// reqHandler := newRequestHandler(in.GetReq())
	// reqHandler.Fulfill()

	// return &cabal.DataResp{Resp: reqHandler.GetResponder()}, nil
	return nil, nil
}

func (s *cabalServer) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {

	rv("-> Unregister Request;")
	rv("   Name=%s; ID=%s; Address=%s", in.GetName(), in.GetId(), in.GetAddress())

	c, err := GetCore()
	if err != nil {
		return &cabal.UnregisterResp{
			Ack:    false,
			Status: &cabal.Status{Sender: "gmbh-core", Target: in.GetId(), Error: "internal server error"},
		}, nil
	}

	service, err := c.Router.LookupService(in.GetName())
	if err != nil {
		return &cabal.UnregisterResp{
			Ack:    false,
			Status: &cabal.Status{Sender: "gmbh-core", Target: in.GetId(), Error: "not found"},
		}, nil
	}

	service.MarkShutdown()

	return &cabal.UnregisterResp{
		Ack: true,
		Status: &cabal.Status{
			Sender: "gmbh-core",
			Target: in.GetId()},
	}, nil
}

func (s *cabalServer) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {
	return &cabal.QueryResponse{}, nil
}

func (s *cabalServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {

	rv("-> Service Registration;")
	rv("   Sender=%s; Target=%s; Status=%s; Action=%s", in.GetSender(), in.GetTarget(), in.GetAction())
	rv("   Message=%s", in.GetMessage())

	return &cabal.ServiceUpdate{Message: "unimp"}, nil
}

func (s *cabalServer) Alive(ctx context.Context, ping *cabal.Ping) (*cabal.Pong, error) {
	return &cabal.Pong{Time: time.Now().Format(time.Stamp)}, nil
}

func rv(msg string, a ...interface{}) {
	notify.LnMagentaF("[rpc] "+msg, a...)
}

func rd(msg string, a ...interface{}) {
	notify.LnCyanF("[data] "+msg, a...)
}

func rve(msg string, a ...interface{}) {
	notify.LnRedF("[rpc] "+msg, a...)
}

// //////////////////////////////////////////////////////////////////////////////////////////
// // Converters
// //////////////////////////////////////////////////////////////////////////////////////////

// // ServiceToRPC translates one service to cabal form
// func ServiceToRPC(s service.Service) *cabal.Service {

// 	r := &cabal.Service{
// 		Id:      s.ID,
// 		Name:    s.Static.Name,
// 		Path:    s.Path,
// 		LogPath: s.Path + defaults.SERVICE_LOG_PATH + defaults.SERVICE_LOG_FILE,
// 	}

// 	if s.Mode == service.Managed {

// 		info := s.Process.GetInfo()

// 		r.Pid = int32(info.PID)
// 		r.Fails = int32(info.Fails)
// 		r.Address = s.Address
// 		r.Restarts = int32(info.Restarts)
// 		r.StartTime = info.StartTime.Format(time.RFC3339)
// 		r.FailTime = info.DeathTime.Format(time.RFC3339)
// 		r.Errors = s.Process.GetErrors()

// 		r.Mode = "managed"
// 		switch s.Process.GetStatus() {
// 		case process.Stable:
// 			r.Status = "Stable"
// 		case process.Running:
// 			r.Status = "Running"
// 		case process.Failed:
// 			r.Status = "Failed"
// 		case process.Killed:
// 			r.Status = "Killed"
// 		}
// 	} else if s.Mode == service.Remote {
// 		r.Mode = "remote"
// 		r.Status = "-"
// 	}
// 	return r
// }

// func serviceToStruct() *service.Service {
// 	return nil
// }
