package rpcconv

import (
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service"
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
		Id:        s.ID,
		Name:      s.Static.Name,
		Path:      s.Path,
		LogPath:   s.Path + defaults.SERVICE_LOG_PATH + defaults.SERVICE_LOG_FILE,
		Pid:       int32(procRuntime.Pid),
		Fails:     int32(procRuntime.Fails),
		Restarts:  int32(procRuntime.Restarts),
		StartTime: procRuntime.StartTime.Format(time.RFC3339),
		FailTime:  procRuntime.DeathTime.Format(time.RFC3339),
		Errors:    s.GetProcess().ReportErrors(),
	}
	if s.Mode == service.Managed {
		rpcService.Mode = "managed"
		if s.Process.GetStatus() {
			errs := s.Process.ReportErrors()
			if len(errs) == 0 {
				rpcService.Status = "running"
			} else {
				rpcService.Status = "degraded"
			}
		} else {
			rpcService.Status = "failed"
		}
	} else {
		rpcService.Mode = "serviceToRPC.nonmanagedServiceError"
		rpcService.Status = "non-managed"
	}
	return rpcService
}

func serviceToStruct() *service.Service {
	return nil
}
