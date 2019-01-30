package process

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

// Process represents the runtime process of a service
type Process struct {
	Control control
	Info    info
	Runtime runtime
	Errors  perr
}

// control represents the interface for which all types of services must adhere.
// This allows them to be controlled by gmbh
type control interface {
	Start(p *Process) (int, error)
	Restart(p *Process) (int, error)
	ForkExec(p *Process, pid chan int)
	GetCmd(p *Process) *exec.Cmd
	HandleFailure(p *Process)
}

type info struct {
	name  string
	args  []string
	env   []string
	path  string
	dir   string
	build bool
}

type runtime struct {
	running     bool
	userKilled  bool
	StartTime   time.Time
	DeathTime   *time.Time
	Pid         int
	numRestarts int
}

type perr struct {
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

func checkDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}
