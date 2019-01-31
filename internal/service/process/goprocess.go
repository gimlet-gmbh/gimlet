package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
)

// GoProc is a fulfilment of the Process interface for GoProcesses
type GoProc struct {
	Inf    *Info
	Run    *Runtime
	Err    *Perr
	Update *sync.Mutex
}

// NewGoProc returns a new process interface fulfilled in go
func NewGoProc(name, path, dir string) *GoProc {
	return &GoProc{
		Inf: &Info{
			name: name,
			args: []string{},
			path: path,
			dir:  dir,
		},
		Run: &Runtime{
			// running:       false,
			// userKilled:    false,
			// userRestarted: false,
			Pid: -1,
			// numRestarts:   0,
		},
		Err: &Perr{
			errors: []error{},
		},
		Update: &sync.Mutex{},
	}
}

// Start a go process
func (g *GoProc) Start() (int, error) {
	getPidChan := make(chan int, 1)
	go g.ForkExec(getPidChan)
	pid := <-getPidChan

	if pid != -1 {
		// p.log(fmt.Sprintf("proccess started pid=(%d)", pid))
		return pid, nil
	}
	return -1, errors.New("could not start process")
}

// Restart a go process
func (g *GoProc) Restart(fromFailed bool) (int, error) {
	if !fromFailed {
		g.Update.Lock()
		g.Run.numRestarts = 0
		g.Update.Unlock()
	}
	return g.Start()
}

// Kill a go process
func (g *GoProc) Kill() {
	g.Update.Lock()
	defer g.Update.Unlock()
	g.Run.running = false
	g.Run.userKilled = false
}

// ForkExec a go process
func (g *GoProc) ForkExec(pid chan int) {

	cmd := g.getCmd()
	listener := make(chan error)
	err := cmd.Start()
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not start process (error=%v)", err))
		pid <- -1
		return
	}

	go func() {
		listener <- cmd.Wait()
	}()

	g.setRuntime(cmd.Process.Pid)
	pid <- cmd.Process.Pid

	select {
	case error := <-listener:
		if err != nil {
			// l.Message("proc error", "err: "+error.Error())
			g.Err.errors = append(g.Err.errors, error)
			// gprint.Err(fmt.Sprintf("Process Failed: %d", p.Runtime.Pid), 0)
		}

		if g.Run.userKilled {
			return
		}

		g.handleFailure()
	}
}

func (g *GoProc) getCmd() *exec.Cmd {
	var cmd *exec.Cmd
	cmd = exec.Command(g.Inf.path, g.Inf.args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	file, err := createLogFile(g.Inf.dir+defaults.SERVICE_LOG_PATH, defaults.SERVICE_LOG_FILE)
	if err != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}
	cmd.Stdout = file
	cmd.Stderr = file
	return cmd
}
func (g *GoProc) handleFailure() {
	g.Update.Lock()
	g.Run.DeathTime = time.Now()
	g.Run.running = false
	if g.Run.userKilled {
		notify.StdMsgErr("user killed")
	} else {
		// Should we restart?
		if g.Run.numRestarts < defaults.NUM_RETRIES {

			msg := fmt.Sprintf("restart attempt %d/%d at %v", g.Run.numRestarts+1, defaults.NUM_RETRIES, time.Now().Format(time.RFC3339))
			notify.StdMsgErr(g.Inf.name + " " + msg)
			g.Err.errors = append(g.Err.errors, errors.New(msg))

			g.Run.numRestarts++
			g.Update.Unlock()
			g.Restart(true)
			return
		}

		msg := fmt.Sprintf("must restart manually: %v", time.Now().Format(time.RFC3339))
		notify.StdMsgErr(g.Inf.name + " " + msg)
		g.Err.errors = append(g.Err.errors, errors.New(msg))

	}
	g.Update.Unlock()
}

// GetStatus of the process
func (g *GoProc) GetStatus() bool {
	return g.Run.running
}

// GetInfo about a go process
func (g *GoProc) GetInfo() *Info {
	return g.Inf
}

// GetRuntime info about a go process
func (g *GoProc) GetRuntime() *Runtime {
	return g.Run
}

// ReportErrors of a go process
func (g *GoProc) ReportErrors() []string {
	ret := []string{}
	for _, e := range g.Err.errors {
		ret = append(ret, e.Error())
	}
	return ret
}

func (g *GoProc) setRuntime(pid int) {
	g.Update.Lock()
	defer g.Update.Unlock()
	g.Run.running = true
	g.Run.StartTime = time.Now()
	g.Run.Pid = pid
}
