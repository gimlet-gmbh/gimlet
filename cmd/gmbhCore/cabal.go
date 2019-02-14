package main

/*
 * cabal.go
 * Implements the gRPC server and client for the gmbhCore Cabal Server
 * Abe Dick
 * Nov 2018
 */

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/gmbh-micro/cabal"
// 	"github.com/gmbh-micro/defaults"
// 	"github.com/gmbh-micro/notify"
// 	"github.com/gmbh-micro/service"
// 	"github.com/gmbh-micro/service/process"
// )

// func v(msg string) {
// 	notify.StdMsgBlueNoPrompt(" [cbl] " + msg)

// }

// // cabalServer is for gRPC interface for the gmbhCore service coms server
// type cabalServer struct{}

// func (s *cabalServer) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {

// 	v(fmt.Sprintf("ephemeral registration service"))
// 	v(fmt.Sprintf("name=%s", in.NewServ.GetName()))
// 	v(fmt.Sprintf("reported mode=%s", in.NewServ.GetMode()))

// 	lookupService, err := core.Router.LookupService(in.NewServ.GetName())
// 	if err != nil {

// 		if err.Error() == "router.LookupService.NotFound" {

// 			v("not found in lookup")

// 			if in.NewServ.GetMode() != cabal.NewService_MANAGED {
// 				v("new service mode is not managed")
// 				lookupService, err = core.registerPlanetaryService(
// 					in.GetNewServ().GetName(),
// 					in.GetNewServ().GetAliases(),
// 					in.GetNewServ().GetIsClient(),
// 					in.GetNewServ().GetIsServer())
// 				if err != nil {
// 					v("cannot add service")
// 					return &cabal.RegServRep{Status: err.Error()}, nil
// 				}
// 			}
// 		} else {
// 			v("other error=" + err.Error())
// 		}
// 	}

// 	if lookupService == nil {
// 		v("service should not be nil")
// 		return &cabal.RegServRep{Status: "error"}, nil
// 	}

// 	v("passed error")
// 	v("retrieved service name=" + lookupService.Static.Name)

// 	if !core.Config.Daemon {
// 		notify.StdMsgMagentaNoPrompt(fmt.Sprintf("[serv] <(%s)- processing ephem-reg request; name=(%s); aliases=(%s); mode=(%s)", lookupService.ID, in.NewServ.GetName(), strings.Join(in.NewServ.GetAliases(), ","), lookupService.GetMode()))
// 		if lookupService.Static.IsServer {
// 			notify.StdMsgMagentaNoPrompt(fmt.Sprintf("       -(%s)> success; address=(%v)", lookupService.ID, lookupService.Address))
// 		} else {
// 			notify.StdMsgMagentaNoPrompt(fmt.Sprintf("       -(%s)> success;", lookupService.ID))
// 		}
// 	}

// 	reply := &cabal.RegServRep{
// 		Status:   "acknowledged",
// 		ID:       lookupService.ID,
// 		Mode:     lookupService.Mode.String(),
// 		CorePath: core.ProjectPath,
// 		Address:  lookupService.Address,
// 	}

// 	return reply, nil
// }

// func (s *cabalServer) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {

// 	reqHandler := newRequestHandler(in.GetReq())
// 	reqHandler.Fulfill()

// 	return &cabal.DataResp{Resp: reqHandler.GetResponder()}, nil
// }

// func (s *cabalServer) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {
// 	reply := &cabal.UnregisterResp{Awk: false}
// 	return reply, nil
// }

// func (s *cabalServer) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {
// 	return nil, nil
// }

// func (s *cabalServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {
// 	return &cabal.ServiceUpdate{Message: "unimp"}, nil
// }

// func (s *cabalServer) Alive(ctx context.Context, ping *cabal.Ping) (*cabal.Pong, error) {
// 	return &cabal.Pong{Time: time.Now().Format(time.Stamp)}, nil
// }

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
