package rpcconv

import (
	"time"

	"github.com/gmbh-micro/cabal"
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
	rpcService := &cabal.Service{
		Id:        s.ID,
		Name:      s.Static.Name,
		Path:      s.Path,
		Pid:       int32(s.Process.GetRuntime().Pid),
		StartTime: s.Process.GetRuntime().StartTime.Format(time.RFC3339),
	}
	if !s.ActiveProcess {
		rpcService.Status = "uninit"
	} else {
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
	}
	return rpcService
}

func serviceToStruct() *service.Service {
	return nil
}
