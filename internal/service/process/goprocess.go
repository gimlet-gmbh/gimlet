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
	status Status
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
		status: Initialized,
		Run: &Runtime{
			Pid: -1,
		},
		Err: &Perr{
			errors: []error{},
		},
		Update: &sync.Mutex{},
	}
}

// Start a go process
func (g *GoProc) Start() (int, error) {

	g.Run.userKilled = false

	getPidChan := make(chan int, 1)
	go g.ForkExec(getPidChan)
	pid := <-getPidChan

	if pid != -1 {
		g.status = Running
		return pid, nil
	}
	g.status = Failed
	return -1, errors.New("GoProc.Start.unableToStartProcess")
}

// Restart a go process
func (g *GoProc) Restart(fromFailed bool) (int, error) {
	if !fromFailed {
		g.Update.Lock()
		g.Run.restartCounter = 0
		g.Run.Restarts++
		g.Run.userKilled = true
		g.Update.Unlock()
		if g.Run.running {
			g.Kill(fromFailed)
			time.Sleep(time.Second * 2)
		}
	}
	return g.Start()
}

// Kill a go process
func (g *GoProc) Kill(withoutRestart bool) {

	g.Update.Lock()
	g.Run.running = false
	if withoutRestart {
		g.Run.userKilled = false
	}
	g.status = Killed
	g.Update.Unlock()

	g.raise(g.Run.Pid, syscall.SIGINT)
}

// ForkExec a go process
func (g *GoProc) ForkExec(pid chan int) {

	cmd := g.getCmd()
	listener := make(chan error)
	err := cmd.Start()
	if err != nil {
		// fmt.Println(fmt.Sprintf("Could not start process (error=%v)", err))
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
			g.Err.errors = append(g.Err.errors, error)
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
	g.status = Failed

	if g.Run.userKilled {
		notify.StdMsgErr("user killed")
	} else {
		// Should we restart?
		reset := false
		// if it has been more than defaults.TIMEOUT seconds since it has last failed
		// restart the counter because it was previously stable
		if time.Since(g.Run.StartTime) > (time.Second * defaults.TIMEOUT) {
			g.Err.errors = append(g.Err.errors, errors.New("process no longer stable"))
			reset = true
			g.Run.restartCounter = 0
		}

		if g.Run.restartCounter < defaults.NUM_RETRIES {
			msg := ""
			if reset {
				msg = fmt.Sprintf("restart attempt at %v with pid=(%d)", time.Now().Format(time.RFC3339), g.Run.Pid)
			} else {
				msg = fmt.Sprintf("restart attempt %d/%d at %v with pid=(%d)", g.Run.restartCounter+1, defaults.NUM_RETRIES, time.Now().Format(time.RFC3339), g.Run.Pid)
			}
			g.Err.errors = append(g.Err.errors, errors.New(msg))

			g.Run.Fails++
			g.Run.restartCounter++
			g.Update.Unlock()
			time.Sleep(time.Second * 2)
			g.Restart(true)
			return
		}

		msg := fmt.Sprintf("must restart manually: %v, last known pid=(%d)", time.Now().Format(time.RFC3339), g.Run.Pid)
		g.Err.errors = append(g.Err.errors, errors.New(msg))

	}
	g.Run.Pid = -1
	g.Update.Unlock()
}

// GetStatus of the process
func (g *GoProc) GetStatus() Status {
	return g.status
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

// raise finds a process by pid and then sends sig to it
func (g *GoProc) raise(pid int, sig os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(sig)
}
