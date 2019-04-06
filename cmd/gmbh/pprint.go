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
	"github.com/gmbh-micro/rpc/intrigue"
)

const report = ` ┌── %s ────────────────────────────────────────────────────
 │ ID		: %s
 │ PID		: %s
 │ Language	: %s
 │ Mode		: %s
 │ Start	: %s
 │ Status	: %s
 │ Address	: %s
 │ Failures	: %d
 │ Restarts	: %d
 │ Log		: %s`

func reportOne(p *intrigue.Service) {
	fmt.Printf(
		report+"\n",
		p.Name,
		p.Id,
		getPid(p.Pid),
		p.Language,
		p.Mode,
		p.StartTime,
		getStatus(p.Status),
		p.Address,
		p.Fails,
		p.Restarts,
		p.LogPath,
	)
	if len(p.Errors) != 0 {
		fmt.Printf(" │ Errors\n")
	}
	for _, e := range p.Errors {
		fmt.Printf(" │ 		: %s\n", e)
	}
}

func formatLine(attr, val, sep string) string {
	return fmt.Sprintf("%-9s %s %s", attr, sep, val)
}

func getBoxTop(name string, length int) string {
	length = length - len(name)
	if length < 0 {
		length = 0
	}
	// \u2510
	return fmt.Sprintf(" \u250C%s %s %s", strings.Repeat("\u2500", 2), name, strings.Repeat("\u2500", length))
}

func getBoxLine(data string) string {
	end := ""
	return fmt.Sprintf(" \u2502 %-38s %s", data, end)
}

func pprintListOne(pm []*intrigue.ProcessManager) {
	for _, r := range pm {
		for _, s := range r.GetServices() {
			reportOne(s)
		}
	}
}

func pprintListAll(remotes []*intrigue.ProcessManager, services []*intrigue.CoreService) {

	m := attachServicesToRemotes(remotes, services)

	// map[remote.ID][coreService.Name]service
	n := attachServicesToMap(remotes)
	for _, r := range remotes {
		fmt.Println(reportRemoteHeader())
		fmt.Println(reportRemotePM(r))
		for _, s := range m[r.ID] {
			fmt.Println(reportRemote(n[r.ID][s.Name], s))
		}
		fmt.Println()
	}
}

func attachServicesToRemotes(remotes []*intrigue.ProcessManager, services []*intrigue.CoreService) map[string][]*intrigue.CoreService {
	m := make(map[string][]*intrigue.CoreService)
	for _, r := range remotes {
		m[r.ID] = make([]*intrigue.CoreService, 0)
	}
	for _, s := range services {
		m[s.ParentID] = append(m[s.ParentID], s)
	}
	return m
}

func attachServicesToMap(remotes []*intrigue.ProcessManager) map[string]map[string]*intrigue.Service {
	m := make(map[string]map[string]*intrigue.Service)
	for _, r := range remotes {
		m[r.ID] = make(map[string]*intrigue.Service)
		for _, s := range r.Services {
			m[r.ID][s.Name] = s
		}
	}
	return m
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
	if len(ts) == 2 {
		return "<1m"
	}
	return ts[:len(ts)-2]
}

func getName(s string) string {
	blue := color.New(color.FgBlue).SprintFunc()
	return blue(fmt.Sprintf("%-12s", s))
}

func reportRemoteHeader() string {
	u := color.New(color.Underline).SprintFunc()
	return u(fmt.Sprintf(" %-9s \u2502 %-8s \u2502 %-6s \u2502 %-7s \u2502 %-3s \u2502 %-12s \u2502 %-4s \u2502 %-20s ",
		"ID",
		"Status",
		"PID",
		"Uptime",
		"Err",
		"Name",
		"Lang",
		"Address",
	))
}

func reportRemote(p *intrigue.Service, c *intrigue.CoreService) string {
	var id, status, up, name, address, language string
	var errs, pid int

	if p != nil {
		id = p.Id
		status = getStatus(p.Status)
		up = getUptime(p.StartTime)
		pid = int(p.Pid)
		errs = len(p.Errors)
		language = p.Language
	}

	if c != nil {
		name = getName(c.Name)
		address = c.Address
	}

	return fmt.Sprintf(" %-9s \u2502 %-8s \u2502 %-6d \u2502 %-7s \u2502 %-3d \u2502 %-12s \u2502 %-4s \u2502 %-20s ",
		id,
		status,
		pid,
		up,
		errs,
		name,
		language,
		address,
	)
}

func reportRemotePM(pm *intrigue.ProcessManager) string {
	return fmt.Sprintf(" %-9s \u2502 %-8s \u2502 %-6s \u2502 %-7s \u2502 %-3d \u2502 %-12s \u2502 %-4s \u2502 %-20s ",
		pm.ID,
		getStatus(pm.Status),
		"-",
		getUptime(pm.StartTime),
		len(pm.Errors),
		"remoteProcm",
		"go",
		pm.Address,
	)
}
