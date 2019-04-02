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

func reportOne(p *intrigue.Service, h string) {
	fmt.Println(h + getBoxTop(p.Name, 54))
	fmt.Println(h + getBoxLine(formatLine("ID", p.Id, ":")))
	if p.Pid != 0 {
		fmt.Println(h + getBoxLine(formatLine("PID", getPid(p.Pid), ":")))
	}
	fmt.Println(h + getBoxLine(formatLine("Mode", p.Mode, ":")))
	fmt.Println(h + getBoxLine(formatLine("Start", p.StartTime, ":")))
	fmt.Println(h + getBoxLine(formatLine("Status", getStatus(p.Status), ":")))
	fmt.Println(h + getBoxLine(formatLine("Address", p.Address, ":")))
	fmt.Println(h + getBoxLine(formatLine("Failures", strconv.Itoa(int(p.Fails)), ":")))
	if p.Fails > 0 {
		fmt.Println(h + getBoxLine(formatLine("Failed At", p.FailTime, ":")))
	}
	fmt.Println(h + getBoxLine(formatLine("Restarts", strconv.Itoa(int(p.Restarts)), ":")))
	if p.Path != "-" {
		fmt.Println(h + getBoxLine(formatLine("Path", p.Path, ":")))
	}
	if p.LogPath != "-" {
		fmt.Println(h + getBoxLine(formatLine("Log", p.LogPath, ":")))
	}
	if len(p.Errors) > 0 {
		fmt.Println(h + getBoxLine(formatLine("Errors", p.Errors[0], ":")))
		for i, e := range p.Errors {
			if i == 0 {
				continue
			}
			fmt.Println(h + getBoxLine(formatLine("", " "+e, "")))
			// if i > 5 {
			// 	fmt.Println(h + getBoxLine(formatLine("", "...query individually for more.", "")))
			// 	break
			// }
		}
	}
	fmt.Println()
}

func reportCluster(c *intrigue.ProcessManager) {
	fmt.Println(getBoxTop(c.ID, 64))
	fmt.Println(getBoxLine(formatLine("Address", c.Address, ":")))
	fmt.Println(getBoxLine(formatLine("Services", "", "")))
	for _, p := range c.Services {
		reportOne(p, " \u2502 ")
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
	// end := "\u2502\n"
	// if len(data) > 38 {
	// 	end = "\n"
	// }
	end := ""
	return fmt.Sprintf(" \u2502 %-38s %s", data, end)
}

func pprintListOne(pm []*intrigue.ProcessManager) {
	for _, r := range pm {
		reportCluster(r)
	}
}

func pprintListAll(remotes []*intrigue.ProcessManager, services []*intrigue.CoreService) {
	fmt.Println(reportRemoteHeader())

	// fmt.Println(remotes)
	// for

	// rs := matchRemotesCores(remotes, services)
	// for _, pm := range remotes {
	// 	fmt.Println(reportRemotePM(pm))
	// 	for _, s := range rs {
	// 		if s.id == pm.ID {
	// 			fmt.Println(reportRemote(s.pm, s.s))
	// 		}
	// 	}
	// }
	// for _, s := range rs {
	// 	if s.id == "" {
	// 		fmt.Println(reportRemote(s.pm, s.s))
	// 	}
	// }
}

func matchRemotesCores(remotes []*intrigue.ProcessManager, services []*intrigue.CoreService) []*remoteservice {
	ret := []*remoteservice{}
	for _, s := range services {
		if s.ParentID == "" {
			ret = append(ret, &remoteservice{s: s})
			continue
		}
		for _, pm := range remotes {
			for _, pms := range pm.GetServices() {
				splitID := strings.Split(pms.Id, "-")
				if s.ParentID == splitID[0] {
					ret = append(ret, &remoteservice{s: s, pm: pms, id: pm.ID})
					continue
				}
			}
		}
	}
	return ret
}

type remoteservice struct {
	pm *intrigue.Service
	s  *intrigue.CoreService
	id string
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

func reportProcessHeader() string {
	u := color.New(color.Underline).SprintFunc()
	return u(fmt.Sprintf(" %-3s \u2502 %5s \u2502 %-8s \u2502 %-7s \u2502 %3s \u2502 %-12s \u2502 %-20s",
		"ID",
		"PID",
		"Status",
		"Uptime",
		"Err",
		"Name",
		"Path",
	))
}

func reportProcess(p *intrigue.Service) string {
	return fmt.Sprintf(" %-3s \u2502 %5s \u2502 %-8s \u2502 %-7s \u2502 %-3d \u2502 %-12s \u2502 %s",
		p.Id,
		getPid(p.Pid),
		getStatus(p.Status),
		getUptime(p.StartTime),
		len(p.Errors),
		getName(p.Name),
		p.Path,
	)
}

func reportRemoteHeader() string {
	u := color.New(color.Underline).SprintFunc()
	return u(fmt.Sprintf(" %-9s \u2502 %-8s \u2502 %-6s \u2502 %-7s \u2502 %-3s \u2502 %-12s \u2502 %-20s ",
		"ID",
		"Status",
		"PID",
		"Uptime",
		"Err",
		"Name",
		"Address",
	))
}

func reportRemote(p *intrigue.Service, c *intrigue.CoreService) string {
	return fmt.Sprintf(" %-9s \u2502 %-8s \u2502 %-6d \u2502 %-7s \u2502 %-3d \u2502 %-12s \u2502 %-20s ",
		p.Id,
		getStatus(p.Status),
		p.Pid,
		getUptime(p.StartTime),
		len(p.Errors),
		getName(p.Name),
		c.Address,
	)
}
func reportRemotePM(pm *intrigue.ProcessManager) string {
	return fmt.Sprintf(" %-9s \u2502 %-8s \u2502 %-6s \u2502 %-7s \u2502 %-3d \u2502 %-12s \u2502 %-20s ",
		pm.ID,
		getStatus(pm.Status),
		"-",
		getUptime(pm.StartTime),
		len(pm.Errors),
		"remoteProcm",
		pm.Address,
	)
}
