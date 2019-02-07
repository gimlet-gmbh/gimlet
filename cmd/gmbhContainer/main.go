package main

/*
 * main.go (gmbhContainer)
 * Abe Dick
 * February 2019
 */

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/service"
	"github.com/gmbh-micro/service/process"
)

var s *service.Service
var id string
var con *rpc.Connection

// First arg should be path to file where gmbh-config can be found

func main() {
	id = "c100"
	notify.SetTag("[gmbh-exp] ")
	notify.StdMsgBlue("gmbh container process manager")

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		notify.StdMsgErr("could not create absolute file path")
	}

	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		notify.StdMsgErr("no path specified")
		os.Exit(1)
	}
	notify.StdMsgBlue("Path: " + path)

	s, err = service.NewManagedService(path)
	s.StartLog(dir+"/gmbh", "process-manager.log")
	if err != nil {
		notify.StdMsgErr("could not create service: " + err.Error())
		os.Exit(1)
	}
	notify.StdMsgBlue("Attempting to start " + s.Static.Name)
	pid, err := s.StartService()
	if err != nil {
		notify.StdMsgErr("could not start service: " + err.Error())
		os.Exit(1)
	}
	notify.StdMsgGreen("service process started with pid=" + pid)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()
	connectToCore(s.Static.Name)
	notify.StdMsgDebug("Blocking until control server interface is implemented")
	<-done
	s.KillProcess()

	fmt.Println("")
	notify.StdMsgMagenta("Recieved shutdown signal")

	os.Exit(0)
}

func connectToCore(name string) {
	client, ctx, can, err := rpc.GetControlRequest("localhost:59997", time.Second*5)
	if err != nil {
		panic(err)
	}
	defer can()

	request := &cabal.ServiceUpdate{
		Sender:  "gmbh-container",
		Target:  "core",
		Message: name,
		Action:  "container.register",
	}
	reply, err := client.UpdateServiceRegistration(ctx, request)
	if err != nil {

		notify.StdMsgErr("could not contact core")
		return

		// panic(err)

	}
	id = reply.GetStatus()
	notify.StdMsgGreen(fmt.Sprintf("reponse from core=(%s); address=(%s)", reply.GetMessage(), reply.GetAction()))
	registerControlServer(reply.GetAction())
}

func registerControlServer(address string) {

	con := rpc.NewRemoteConnection()
	con.Address = address
	con.Remote = &remoteServer{}
	err := con.Connect()
	if err != nil {
		notify.StdMsgErr("Error starting remote server=" + err.Error())
		return
	}
	notify.StdMsgGreen("started remote server")
}

type remoteServer struct{}

func (c *remoteServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {
	// notify.StdMsgBlue(fmt.Sprintf("-> Update Service Request; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))
	return &cabal.ServiceUpdate{Message: "unimp"}, nil
}

func (c *remoteServer) RequestRemoteAction(ctx context.Context, in *cabal.Action) (*cabal.Action, error) {
	notify.StdMsgBlue(fmt.Sprintf("-> Request Remote Action; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))

	if in.GetAction() == "request.info" {

		procRuntime := s.GetProcess().GetRuntime()

		si := &cabal.Service{
			Id:        id + "-" + s.ID,
			Name:      s.Static.Name,
			Path:      "-",
			LogPath:   "-",
			Pid:       0,
			Fails:     int32(procRuntime.Fails),
			Restarts:  int32(procRuntime.Restarts),
			StartTime: procRuntime.StartTime.Format(time.RFC3339),
			FailTime:  procRuntime.DeathTime.Format(time.RFC3339),
			Errors:    s.GetProcess().ReportErrors(),
			Mode:      "remote",
		}

		switch s.Process.GetStatus() {
		case process.Stable:
			si.Status = "Stable"
		case process.Running:
			si.Status = "Running"
		case process.Degraded:
			si.Status = "Degraded"
		case process.Failed:
			si.Status = "Failed"
		case process.Killed:
			si.Status = "Killed"
		case process.Initialized:
			si.Status = "Initialized"
		}

		response := &cabal.Action{
			Sender:      s.Static.Name,
			Target:      "gmbh-core",
			Message:     "response.info",
			ServiceInfo: si,
		}
		return response, nil
	}
	return &cabal.Action{Message: "unimp"}, nil

}
