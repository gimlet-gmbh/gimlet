package main

/*
 * main.go (gmbhContainer)
 * Abe Dick
 * February 2019
 */

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/service"
	"github.com/gmbh-micro/service/process"
)

type container struct {
	serv      *service.Service
	con       *rpc.Connection
	to        time.Duration
	mu        *sync.Mutex
	coreAddr  string
	closed    bool
	id        string
	forkError error

	configPath *string
	managed    *bool
	embedded   *bool
	daemon     *bool
}

func startContainer() {

	c.mu = &sync.Mutex{}
	c.con = rpc.NewRemoteConnection("", &remoteServer{})
	c.coreAddr = "localhost:59997"
	c.closed = false
	c.to = time.Second * 5

	if !*c.daemon {
		notify.SetTag("[gmbh-pm] ")
		notify.StdMsg("gmbh container process manager")
	} else {
		notify.SetVerbose(false)
	}

	if *c.configPath == "" {
		notify.StdMsgErr("must specify a config file")
		os.Exit(1)
	}

	run()

}

func run() {
	var err error
	c.serv, err = service.NewManagedService(*c.configPath)
	if err != nil {
		notify.StdMsgErr("could not start service; err=(" + err.Error() + ")")
		os.Exit(1)
	}

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	c.serv.StartLog(dir+"/gmbh", "process-manager.log")

	pid, err := c.serv.StartService()
	if err != nil {
		notify.StdMsgErr("could not start service, error=(" + err.Error() + ")")
		c.forkError = err
	} else {
		notify.StdMsgGreen("started process; pid=(" + pid + ")")
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()

	if *c.managed {
		go connect()
	}

	<-done
	fmt.Println()

	c.serv.KillProcess()

	if *c.managed {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()

		disconnect()
	}

	notify.StdMsg("shutdown signal")
	return
}

func connect() {
	notify.StdMsg("connecting to gmbh-core")

	addr, status := makeConnectRequest()
	for status != nil {
		if status.Error() != "makeConnectRequest.fail" {
			notify.StdMsg("gmbh internal error")
			return
		}

		if c.closed {
			return
		}

		notify.StdMsg("could not connect; retry=(" + c.to.String() + ")")
		time.Sleep(c.to)
		addr, status = makeConnectRequest()
	}

	if addr == "" {
		notify.StdMsg("gmbh internal error, no address returned from core")
		return
	}

	c.con.SetAddress(addr)
	c.con.Remote = &remoteServer{}
	err := c.con.Connect()
	if err != nil {
		notify.StdMsgErr("gmbh connection error=(" + err.Error() + ")")
		return
	}

	notify.StdMsgGreen("connected; address=(" + addr + ")")

}

func disconnect() {
	notify.StdMsg("disconnected")
	c.con.Disconnect()
	c.con.Server = nil
	if !c.closed {
		time.Sleep(c.to)
	}
}

func makeConnectRequest() (string, error) {
	client, ctx, can, err := rpc.GetControlRequest(c.coreAddr, time.Second*5)
	if err != nil {
		panic(err)
	}
	defer can()

	request := &cabal.ServiceUpdate{
		Sender:  "gmbh-container",
		Target:  "core",
		Message: c.serv.Static.Name,
		Action:  "container.register",
	}

	reply, err := client.UpdateServiceRegistration(ctx, request)
	if err != nil {
		notify.StdMsgErr("updateServiceRegistration err=(" + err.Error() + ")")
		return "", errors.New("makeConnectRequest.fail")
	}

	c.id = reply.GetStatus()

	return reply.GetAction(), nil
}

type remoteServer struct{}

func (r *remoteServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {
	notify.StdMsgBlue(fmt.Sprintf("-> Update Service Request; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))

	if in.GetAction() == "core.shutdown" {
		response := &cabal.ServiceUpdate{
			Sender:  c.serv.Static.Name,
			Target:  "gmbh-core",
			Message: "ack",
		}
		if !c.closed {
			go func() {
				disconnect()
				connect()
			}()
		}
		return response, nil
	}

	return &cabal.ServiceUpdate{Message: "unimp"}, nil
}

func (r *remoteServer) RequestRemoteAction(ctx context.Context, in *cabal.Action) (*cabal.Action, error) {
	notify.StdMsgBlue(fmt.Sprintf("-> Request Remote Action; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))

	if in.GetAction() == "request.info" {
		response := &cabal.Action{
			Sender:      c.serv.Static.Name,
			Target:      "gmbh-core",
			Message:     "response.info",
			ServiceInfo: serviceToRPC(c.serv),
		}
		return response, nil
	} else if in.GetAction() == "service.restart" {
		c.serv.RestartProcess()
		response := &cabal.Action{
			Sender:  c.serv.Static.Name,
			Target:  "gmbh-core",
			Message: "action.completed",
		}
		return response, nil
	}
	return &cabal.Action{Message: "unimp"}, nil

}

func serviceToRPC(s *service.Service) *cabal.Service {
	procRuntime := c.serv.GetProcess().GetRuntime()

	si := &cabal.Service{
		Id:        c.id + "-" + c.serv.ID,
		Name:      c.serv.Static.Name,
		Path:      "-",
		LogPath:   "-",
		Pid:       0,
		Fails:     int32(procRuntime.Fails),
		Restarts:  int32(procRuntime.Restarts),
		StartTime: procRuntime.StartTime.Format(time.RFC3339),
		FailTime:  procRuntime.DeathTime.Format(time.RFC3339),
		Errors:    c.serv.GetProcess().ReportErrors(),
		Mode:      "remote",
	}

	switch c.serv.Process.GetStatus() {
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
	return si
}
