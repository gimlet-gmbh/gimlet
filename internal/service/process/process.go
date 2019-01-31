package process

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

// Things to think about
// How will untimely process death be notifed

// Process represents the runtime process of a service
type Process interface {
	Start() (int, error)
	Kill()
	Restart(fromFailed bool) (int, error)
	ForkExec(pid chan int)
	getCmd() *exec.Cmd
	handleFailure()
	GetStatus() bool
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
	running       bool
	userKilled    bool
	userRestarted bool
	StartTime     time.Time
	DeathTime     time.Time
	Pid           int
	numRestarts   int
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

// Quick Implement
// type GoProc struct {
// 	Inf Info
// 	Run Runtime
// 	Err Perr
// }
// func NewGoProc() *GoProc {
// 	return &GoProc{}
// }
// func (g *GoProc) Start() (int, error) {
// 	return -1, nil
// }
// func (g *GoProc) Restart() (int, error) {
// 	return -1, nil
// }
// func (g *GoProc) ForkExec(pid chan int) {
// }
// func (g *GoProc) GetCmd() *exec.Cmd {
// 	return nil
// }
// func (g *GoProc) HandleFailure() {
// }
// func (g *GoProc) GetInfo() Info {
// 	return g.Inf
// }
// func (g *GoProc) GetRuntime() Runtime {
// 	return g.Run
// }
// func (g *GoProc) GetError() Perr {
// 	return g.Err
// }
