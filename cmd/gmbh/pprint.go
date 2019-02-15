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
)

func reportOne(p *cabal.Service, h string) {
	fmt.Println(h + getBoxTop(p.Name, 38))
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
			fmt.Println(h + getBoxLine(formatLine("", e, "")))
			if i > 5 {
				fmt.Println(h + getBoxLine(formatLine("", "...query individually for more.", "")))
				break
			}
		}
	}
	fmt.Println()
}

func reportCluster(c *cabal.ProcessManager) {
	fmt.Println(getBoxTop(c.ID, 42))
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
	return fmt.Sprintf(" \u250C%s %s %s\u2510", strings.Repeat("\u2500", 2), name, strings.Repeat("\u2500", length))
}

func getBoxLine(data string) string {
	// end := "\u2502\n"
	// if len(data) > 38 {
	// 	end = "\n"
	// }
	end := ""
	return fmt.Sprintf(" \u2502 %-38s %s", data, end)
}

func pprintListOne(pm []*cabal.ProcessManager) {
	for _, r := range pm {
		reportCluster(r)
	}
}

func pprintListAll(remote []*cabal.ProcessManager) {
	fmt.Println(reportRemoteHeader())
	for _, p := range remote {
		for _, s := range p.GetServices() {
			fmt.Println(reportRemote(s, p.ID))
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

func reportProcess(p *cabal.Service) string {
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
	return u(fmt.Sprintf(" %-8s \u2502 %-8s \u2502 %-7s \u2502 %-3s \u2502 %-12s ",
		"ID",
		"Status",
		"Uptime",
		"Err",
		"Name",
	))
}

func reportRemote(p *cabal.Service, nid string) string {
	return fmt.Sprintf(" %-8s \u2502 %-8s \u2502 %-7s \u2502 %-3d \u2502 %-12s ",
		p.Id,
		getStatus(p.Status),
		getUptime(p.StartTime),
		len(p.Errors),
		getName(p.Name),
	)
}
