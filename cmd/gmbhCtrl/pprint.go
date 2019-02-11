package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
)

func pprintListOne(service cabal.Service) {
	notify.StdMsg("Displaying service information")
	notify.StdMsgNoPrompt(" ID:\t"+service.GetId(), 2)
	notify.StdMsgNoPrompt(" Name:\t"+service.GetName(), 2)
	notify.StdMsgNoPrompt(" Mode:\t"+service.GetMode(), 2)
	notify.StdMsgNoPrompt(" PID:\t"+strconv.Itoa(int(service.GetPid())), 2)
	notify.StdMsgNoPrompt(" Start:\t"+service.GetStartTime(), 2)
	if service.GetStatus() == "Failed" {
		notify.StdMsgErrNoPrompt(" Status:\t"+service.GetStatus(), 2)
		notify.StdMsgErrNoPrompt(" Failed:\t"+service.GetFailTime(), 2)
	} else if service.GetStatus() == "degraded" {
		notify.StdMsgErrNoPrompt(" Status:\t"+service.GetStatus(), 2)
	} else {
		notify.StdMsgGreenNoPrompt(" Status:\t"+service.GetStatus(), 2)
	}
	notify.StdMsgNoPrompt(" Fails:\t"+strconv.Itoa(int(service.GetFails())), 2)
	notify.StdMsgNoPrompt(" Restarts:"+strconv.Itoa(int(service.GetRestarts())), 2)
	notify.StdMsgNoPrompt(" Path:\t"+service.GetPath(), 2)
	notify.StdMsgNoPrompt(" Logs:\t"+service.GetLogPath(), 2)
	errs := service.GetErrors()
	if len(errs) <= 1 {
		notify.StdMsgNoPrompt(" Errors:\t"+strings.Join(errs, ","), 2)
	} else {
		notify.StdMsgNoPrompt(" Errors:\t"+errs[0], 2)
		errs = errs[1:]
		for _, e := range errs {
			notify.StdMsgNoPrompt("        \t"+e, 2)
		}
	}
}

func pprintListAll(managed []*cabal.Service, remote []*cabal.ProcessManager, planetary []*cabal.Service) {

	if len(managed) != 0 {
		fmt.Println("Managed Services")
		fmt.Println(reportProcessHeader())
		for _, s := range managed {
			fmt.Println(reportProcess(s))
		}
	}

	if len(remote) != 0 {
		fmt.Println("\nRemote Services")
		fmt.Println(reportRemoteHeader())
		for _, p := range remote {
			for _, s := range p.GetServices() {
				fmt.Println(reportRemote(s, p.ID))
			}
		}
	}

	if len(planetary) != 0 {
		notify.StdMsgNoPrompt("\nPlanetary Services")
		for _, p := range planetary {
			fmt.Println(reportRemoteHeader())
			fmt.Println(reportRemote(p, "----"))
		}
	}
}

func getStatus(s string) string {
	status := ""
	if s == "Stable" || s == "Running" {
		green := color.New(color.FgGreen).SprintFunc()
		status = green(fmt.Sprintf("%-8s", s))
	} else if s == "Degraded" {
		yellow := color.New(color.FgYellow).SprintFunc()
		status = yellow(fmt.Sprintf("%-8s", s))
	} else {
		red := color.New(color.FgRed).SprintFunc()
		status = red(fmt.Sprintf("%-8s", s))
	}
	return status
}

func getPid(s int32) string {
	if s == -1 {
		return "-----"
	}
	return strconv.Itoa(int(s))
}

func getUptime(t string) string {
	pt, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return ""
	}
	ts := time.Now().Sub(pt).Truncate(time.Minute).String()
	return ts[:len(ts)-2]
}

func getName(s string) string {
	blue := color.New(color.FgBlue).SprintFunc()
	return blue(fmt.Sprintf("%-12s", s))
}

func reportProcessHeader() string {
	u := color.New(color.Underline).SprintFunc()
	return u(fmt.Sprintf(" %-3s \u2502 %5s \u2502 %-8s \u2502 %-7s \u2502 %-12s \u2502 %-20s",
		"ID",
		"PID",
		"Status",
		"Uptime",
		"Name",
		"Path",
	))
}

func reportProcess(p *cabal.Service) string {
	return fmt.Sprintf(" %-3s \u2502 %5s \u2502 %-8s \u2502 %-7s \u2502 %-12s \u2502 %s",
		p.Id,
		getPid(p.Pid),
		getStatus(p.Status),
		getUptime(p.StartTime),
		getName(p.Name),
		p.Path,
	)
}

func reportRemoteHeader() string {
	u := color.New(color.Underline).SprintFunc()
	return u(fmt.Sprintf(" %-4s \u2502 %-3s \u2502 %-8s \u2502 %-7s \u2502 %-5s \u2502 %-12s ",
		"NID",
		"ID",
		"Status",
		"Uptime",
		"Err",
		"Name",
	))
}

func reportRemote(p *cabal.Service, nid string) string {
	return fmt.Sprintf(" %-4s \u2502 %-3s \u2502 %-8s \u2502 %-7s \u2502 %-5d \u2502 %-12s ",
		nid,
		p.Id,
		getStatus(p.Status),
		getUptime(p.StartTime),
		len(p.Errors),
		getName(p.Name),
	)
}
