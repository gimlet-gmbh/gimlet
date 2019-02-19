package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// LocalManager ; a process manager
type LocalManager struct {
	name string
	args []string
	env  []string
	path string
	dir  string
	// userKilled       bool
	userRestarted    bool
	restartCounter   int
	gracefulshutdown bool
	// true is sent to this buffer when a process has been marked to restart
	// by the user
	restartBuffer chan bool
	// true is sent to this buffer when the command for a process has finished
	exitedBuffer chan bool
	ssignal      syscall.Signal
	logFile      *os.File
	mu           *sync.Mutex
	info         Info
}

// LocalProcessConfig is used to pass relevant data to the process launcher
type LocalProcessConfig struct {
	Name   string
	Path   string
	Dir    string
	LogF   *os.File
	Args   []string
	Env    []string
	Signal syscall.Signal
}

// NewLocalBinaryManager ; as in new process manager to monitor a binary forked from the shell
func NewLocalBinaryManager(conf *LocalProcessConfig) *LocalManager {
	return &LocalManager{
		name:          conf.Name,
		args:          conf.Args,
		env:           conf.Env,
		path:          conf.Path,
		dir:           conf.Dir,
		ssignal:       conf.Signal,
		restartBuffer: make(chan bool, 100),
		exitedBuffer:  make(chan bool, 100),
		mu:            &sync.Mutex{},
		logFile:       conf.LogF,
		info: Info{
			Type:   Binary,
			Status: Initialized,
			Errors: make([]error, 0),
		},
	}
}

// // NewLocalGoManager ; as in new process manager to monitor bulding from go source code
// func NewLocalGoManager(name string, path string, dir string, args []string, env []string) *LocalManager {
// 	return &LocalManager{
// 		// name: name,
// 		args: args,
// 		env:  env,
// 		path: path,
// 		dir:  dir,
// 		mu:   &sync.Mutex{},
// 		info: Info{
// 			Type:   Go,
// 			Errors: make([]error, 0),
// 		},
// 	}
// }

// Start a process if possible
func (m *LocalManager) Start() (PID int, err error) {

	getPidChan := make(chan int, 1)
	getErrChan := make(chan error, 1)
	go m.forkExec(getPidChan, getErrChan)
	pid := <-getPidChan
	if pid != -1 {
		m.mu.Lock()
		m.info.Status = Running
		m.mu.Unlock()
		go m.upgrade()
		return pid, nil
	}
	perr := <-getErrChan
	return -1, fmt.Errorf("[%s] StartError=%s", m.name, perr.Error())
}

// Restart a process.
func (m *LocalManager) Restart(fromFailed bool) (pid int, err error) {

	if m.info.Status == Restarting {
		return -1, fmt.Errorf("cannot restart a process that is already restarting")
	}

	if m.info.Status == Initialized {
		return -1, fmt.Errorf("cannot restart a process that has only been initialized")
	}

	m.mu.Lock()
	previousState := m.info.Status
	m.info.Status = Restarting
	m.mu.Unlock()

	if !fromFailed {
		m.mu.Lock()
		m.restartBuffer <- true
		m.restartCounter = 0
		m.info.Restarts++
		m.mu.Unlock()

		if previousState == Running || previousState == Stable {
			m.Kill(fromFailed)
			fmt.Println("blocking at exit buffer")
			_ = <-m.exitedBuffer
			fmt.Println("value received in exit buffer, returning restart")
		}
	}
	return m.Start()
}

// Kill a process
func (m *LocalManager) Kill(withoutRestart bool) {
	if withoutRestart {
		m.mu.Lock()
		m.info.Status = Killed
		m.mu.Unlock()
	}
	m.raise(m.info.PID, m.ssignal)
}

// GetErrors from local manager
func (m *LocalManager) GetErrors() []string {
	ret := []string{}
	for _, s := range m.info.Errors {
		ret = append(ret, s.Error())
	}
	return ret
}

// GetInfo of the local process
func (m *LocalManager) GetInfo() Info {
	return m.info
}

// GetStatus of the local process
func (m *LocalManager) GetStatus() Status {
	return m.info.Status
}

func (m *LocalManager) forkExec(pid chan int, errChan chan error) {

	cmd := m.getCmd()
	err := cmd.Start()
	if err != nil {
		m.info.Errors = append(m.info.Errors, err)
		errChan <- err
		pid <- -1
		return
	}

	m.mu.Lock()
	m.info.Status = Running
	m.info.StartTime = time.Now()
	m.info.PID = cmd.Process.Pid
	m.mu.Unlock()

	// send the pid back to the caller who is waiting at the channel
	pid <- cmd.Process.Pid

	// wait for the command to finish; could be error, could be nill.
	fmt.Println("waiting at listener")
	err = cmd.Wait()
	fmt.Println("command finished")
	m.exitedBuffer <- true

	if m.gracefulshutdown {
		m.info.Errors = append(m.info.Errors, errors.New("marked for graceful shutdown"))
		return
	}

	fmt.Println("status here=", m.info.Status)
	if m.info.Status == Killed {
		return
	}

	select {

	// case in which there is a value in the restart buffer, as of which
	// we would want to ignore the command exiting with the listener
	case <-m.restartBuffer:
		fmt.Println("restart buffer has value, return")
		return

	// case in which there is no value in the restart buffer, can assume
	// that the service has failed
	default:
		fmt.Println("restart buffer empty, handle failure")

		// Record the error
		if err != nil {
			m.info.Errors = append(m.info.Errors, err)
		}
		m.handleFailure()
	}
}

func (m *LocalManager) getCmd() *exec.Cmd {
	if m.info.Type == Binary {

		var cmd *exec.Cmd

		cmd = exec.Command(m.path, m.args...)
		cmd.Env = m.env
		cmd.Dir = m.dir

		if m.logFile != nil {
			cmd.Stdout = m.logFile
			cmd.Stderr = m.logFile
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		return cmd
	}
	return nil
}

func (m *LocalManager) handleFailure() {
	m.mu.Lock()

	m.info.DeathTime = time.Now()
	m.info.Status = Failed

	// Only give up restarting if the process has beeen attempting to restart n times
	// in the last 30 seconds, if it has been longer than 3 seconds, clear the restart counter

	if time.Since(m.info.StartTime) > time.Second*30 {
		m.restartCounter = 0
	}

	if m.restartCounter < 3 {
		m.info.Errors = append(m.info.Errors, fmt.Errorf("restart=%d/3; time=%s; last-pid=%d", m.restartCounter+1, time.Now().Format(time.Stamp), m.info.PID))
		m.info.Fails++
		m.restartCounter++
		m.mu.Unlock()
		m.Restart(true)

		return
	}

	m.info.Errors = append(m.info.Errors, fmt.Errorf("exceeded restart counter; time=%s; last-pid=%d", time.Now().Format(time.Stamp), m.info.PID))
	m.info.PID = -1
	m.mu.Unlock()
}

// upgrade from Running to Stable if the process.Status is still Running after 30 seconds
func (m *LocalManager) upgrade() {
	for {
		time.Sleep(time.Second * 31)
		if m.info.Status == Running {
			if time.Since(m.info.StartTime) > time.Second*30 {
				m.mu.Lock()
				m.info.Status = Stable
				m.mu.Unlock()
			}
		} else if m.info.Status == Failed {
			return
		}
	}
}

// raise finds a process by pid and then sends sig to it
func (m *LocalManager) raise(pid int, sig os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(sig)
}

// GracefulShutdown blocks restart on failure
func (m *LocalManager) GracefulShutdown() {
	m.mu.Lock()
	m.gracefulshutdown = true
	m.mu.Unlock()
}
