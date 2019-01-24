package pmgmt

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

/*
 * process.go
 * Abe Dick
 * January 2019
 */

// ProcessController represents the interface for which all types of services must adhere
type ProcessController interface {
	Start(p *Process) (int, error)
	Restart(p *Process) (int, error)
	ForkExec(p *Process, pid chan int)
	GetCmd(p *Process) *exec.Cmd
	HandleFailure(p *Process)
}

// Process represents the runtime controller and handling functions
type Process struct {
	ID         int
	Controller ProcessController
	Info       pInfo
	Runtime    pRuntime
	Errs       pError
	// files   pFiles
}

type pInfo struct {
	name  string
	args  []string
	env   []string
	path  string
	dir   string
	build bool
}

type pRuntime struct {
	running       bool
	userKilled    bool
	startTime     time.Time
	lastAliveTime time.Time
	Pid           int
	numRestarts   int
}

type pError struct {
	errors []error
}

func createLogFile(path, filename string) (*os.File, error) {

	checkDir(path)

	stdOut, err := os.OpenFile(path+filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		// gprint.Err(err.Error(), 3)
		return nil, errors.New("could not create log file")
	}

	return stdOut, nil
}

func checkDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// type pFiles struct {
// 	log        *bufio.Writer
// 	stdOut     *os.File
// 	writeMutex sync.Mutex
// }

// func newProcess(name string, path string, build bool) *process {
// 	p := process{
// 		// id: getProcID(),
// 		info: pInfo{
// 			name:  name,
// 			args:  []string{},
// 			path:  path,
// 			build: build,
// 		},
// 		runtime: pRuntime{
// 			running:     false,
// 			userKilled:  false,
// 			pid:         -1,
// 			numRestarts: 0,
// 		},
// 		errs: pError{
// 			errors: []error{},
// 		},
// 	}
// 	// p.files.log = getLogFile(name, p.id)
// 	// p.files.stdOut = getStdOutFile(name, p.id)

// 	// if p.files.log != nil {
// 	// 	p.log("process created")
// 	// }

// 	return &p
// }

// func (p *process) start() (int, error) {

// 	getPidChan := make(chan int, 1)
// 	go p.forkExec(getPidChan)
// 	pid := <-getPidChan

// 	if pid != -1 {
// 		p.log(fmt.Sprintf("proccess started pid=(%d)", pid))
// 		return pid, nil
// 	}
// 	return -1, errors.New("could not start process")
// }

// func (p *process) restart(user bool) (int, error) {
// 	p.kill(true)
// 	if !user {
// 		p.runtime.numRestarts++
// 		// p.log(fmt.Sprintf("attempting to restart process, attempt=(%d/%d)", p.runtime.numRestarts, c.PList.Retries))
// 	} else {
// 		p.runtime.numRestarts = 0
// 		p.log("user initiated process restart")
// 	}
// 	return p.start()
// }

// func (p *process) kill(restart bool) error {
// 	err := syscall.Kill(-p.runtime.pid, syscall.SIGKILL)
// 	if err != nil {
// 		msg := fmt.Sprintf("could not kill process: %v", err)
// 		p.log(msg)
// 		return errors.New(msg)
// 	}
// 	return nil
// }

// func (p *process) forkExec(pid chan int) {

// 	cmd := p.getCmd()
// 	listener := make(chan error)
// 	err := cmd.Start()
// 	if err != nil {
// 		p.log(fmt.Sprintf("Could not start process (id=%d) (error=%v)", p.id, err))
// 		pid <- -1
// 		return
// 	}

// 	go func() {
// 		listener <- cmd.Wait()
// 	}()

// 	p.setRuntime(cmd.Process.Pid)
// 	pid <- cmd.Process.Pid

// 	select {
// 	case error := <-listener:
// 		if err != nil {
// 			// l.Message("proc error", "err: "+error.Error())
// 			p.errs.errors = append(p.errs.errors, error)
// 		}

// 		if p.runtime.userKilled {
// 			return
// 		}

// 		p.handleFailure()
// 	}
// }

// func (p *process) handleFailure() {
// 	p.log(fmt.Sprintf("process failure reported, pid=(%d)", p.runtime.pid))
// 	p.runtime.pid = -1
// 	p.runtime.running = false
// 	// if p.runtime.numRestarts < c.PList.Retries {
// 	// 	p.restart(false)
// 	// } else {
// 	// 	p.log(fmt.Sprintf("exceeded restart count, must manually restart using proc.id (%d/%d)", p.runtime.numRestarts, c.PList.Retries))
// 	// }
// }

// func (p *process) getCmd() *exec.Cmd {
// 	var cmd *exec.Cmd
// 	if p.info.build {
// 		cmd = exec.Command(p.info.path, p.info.args...)
// 	} else {
// 		cmd = exec.Command("go", p.info.args...)
// 		cmd.Dir = p.info.path
// 	}
// 	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
// 	cmd.Stdout = p.files.stdOut
// 	cmd.Stderr = p.files.stdOut
// 	return cmd
// }

// func (p *process) setRuntime(pid int) {
// 	p.runtime.running = true
// 	p.runtime.startTime = time.Now()
// 	p.runtime.lastAliveTime = time.Now()
// 	p.runtime.pid = pid
// }

// func (p *process) setArgs() error {
// 	// goFiles, err := getGoFilesInDir(p.info.path + "/")
// 	// if err != nil || len(goFiles) == 0 {
// 	// 	return errors.New("could not find go files in /*.go")
// 	// }
// 	// p.info.args = []string{"run"}
// 	// p.info.args = append(p.info.args, goFiles...)
// 	return nil
// }

// // log - logs (msg) to the log file associated with this process
// func (p *process) log(msg string) {
// 	if p.files.log != nil {
// 		outB, _ := json.Marshal(time.Now().Format("2006-01-02 15:04:05") + " " + msg)
// 		p.files.writeMutex.Lock()
// 		p.files.log.Write(outB)
// 		p.files.log.WriteString("\n")
// 		p.files.log.Flush()
// 		p.files.writeMutex.Unlock()
// 	}
// }
