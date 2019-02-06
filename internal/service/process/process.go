package process

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

// Things to think about
// How will untimely process death be notifed

// Status is the enumerated stated of the current status of the service
type Status int

const (
	// Stable ; the process is running and without error for x amunt of time
	Stable Status = 1

	// Running ; the process is running but not yet stable
	Running Status = 2

	// Degraded ; the process is running but with error
	Degraded Status = 3

	// Failed ; the process has ran and failed
	Failed Status = 4

	// Killed ; the process has been killed
	Killed Status = 5

	// Initialized ; the process object has been instantiated but not started
	Initialized Status = 6

	// Invalid ; the service is running in planetary mode
	Invalid Status = 7
)

// Process represents the runtime process of a service
type Process interface {
	Start() (int, error)
	Kill(withoutRestart bool)
	Restart(fromFailed bool) (int, error)
	ForkExec(pid chan int)
	getCmd() *exec.Cmd
	handleFailure()
	GetStatus() Status
	GetInfo() *Info
	GetRuntime() *Runtime
	ReportErrors() []string
}

// Info stores static info about processes
type Info struct {
	name  string
	args  []string
	env   []string
	path  string
	dir   string
	build bool
}

// Runtime stores runtime info about processes
type Runtime struct {
	running        bool
	userKilled     bool
	userRestarted  bool
	StartTime      time.Time
	DeathTime      time.Time
	Pid            int
	Restarts       int
	restartCounter int
	Fails          int
}

// Perr stores error information about processes
type Perr struct {
	errors []error
}

func createLogFile(path, filename string) (*os.File, error) {

	checkDir(path)

	stdOut, err := os.OpenFile(path+filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return nil, errors.New("could not create log file")
	}

	return stdOut, nil
}

// raise finds a process by pid and then sends sig to it
func raise(pid int, sig os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(sig)
}

func checkDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// EmptyProc fulfills process interface without holding any data. This is used
// for processes that are not in managed mode
type EmptyProc struct {
	Inf    *Info
	status Status
	Run    *Runtime
	Err    Perr
}

// NewEmptyProc returns emptyProc
func NewEmptyProc() *EmptyProc {
	return &EmptyProc{}
}

// Start placeholder
func (e *EmptyProc) Start() (int, error) {
	return -1, errors.New("emptyProcess.notImplemented")
}

// Kill placeholder
func (e *EmptyProc) Kill(withoutRestart bool) {
}

// Restart placeholder
func (e *EmptyProc) Restart(fromFailed bool) (int, error) {
	return -1, errors.New("emptyProcess.notImplemented")
}

// ForkExec placeholder
func (e *EmptyProc) ForkExec(pid chan int) {
}

func (e *EmptyProc) getCmd() *exec.Cmd {
	return nil
}

func (e *EmptyProc) handleFailure() {
}

// GetStatus placeholder
func (e *EmptyProc) GetStatus() Status {
	return Invalid
}

// GetInfo placeholder
func (e *EmptyProc) GetInfo() *Info {
	return nil
}

// GetRuntime placeholder
func (e *EmptyProc) GetRuntime() *Runtime {
	return nil
}

// ReportErrors placeholder
func (e *EmptyProc) ReportErrors() []string {
	return []string{"emptyProcess.notImplemented"}
}
