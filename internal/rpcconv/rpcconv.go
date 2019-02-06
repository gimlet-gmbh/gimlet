package rpcconv

import (
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service"
	"github.com/gmbh-micro/service/process"
)

// ServicesToRPCs translates an array of service pointers to an array of cabal service pointers
func ServicesToRPCs(ss []*service.Service) []*cabal.Service {
	ret := []*cabal.Service{}
	for _, s := range ss {
		ret = append(ret, ServiceToRPC(*s))
	}
	return ret
}

// ServiceToRPC translates one service to cabal form
func ServiceToRPC(s service.Service) *cabal.Service {

	procRuntime := s.GetProcess().GetRuntime()

	rpcService := &cabal.Service{
		Id:      s.ID,
		Name:    s.Static.Name,
		Path:    s.Path,
		LogPath: s.Path + defaults.SERVICE_LOG_PATH + defaults.SERVICE_LOG_FILE,
	}

	if s.Mode == service.Managed {

		rpcService.Pid = int32(procRuntime.Pid)
		rpcService.Fails = int32(procRuntime.Fails)
		rpcService.Restarts = int32(procRuntime.Restarts)
		rpcService.StartTime = procRuntime.StartTime.Format(time.RFC3339)
		rpcService.FailTime = procRuntime.DeathTime.Format(time.RFC3339)
		rpcService.Errors = s.GetProcess().ReportErrors()

		rpcService.Mode = "managed"
		switch s.Process.GetStatus() {
		case process.Stable:
			rpcService.Status = "Stable"
		case process.Running:
			rpcService.Status = "Running"
		case process.Degraded:
			rpcService.Status = "Degraded"
		case process.Failed:
			rpcService.Status = "Failed"
		case process.Killed:
			rpcService.Status = "Killed"
		case process.Initialized:
			rpcService.Status = "Initialized"
		}
	} else if s.Mode == service.Remote {
		rpcService.Mode = "remote"
		rpcService.Status = "-"
	}
	return rpcService
}

func serviceToStruct() *service.Service {
	return nil
}
