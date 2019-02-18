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
	args             []string
	env              []string
	path             string
	dir              string
	userKilled       bool
	userRestarted    bool
	restartCounter   int
	gracefulshutdown bool
	ssignal          syscall.Signal
	logFile          *os.File
	mu               *sync.Mutex
	info             Info
}

// LocalProcessConfig is used to pass relevant data to the process launcher
type LocalProcessConfig struct {
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
		args:    conf.Args,
		env:     conf.Env,
		path:    conf.Path,
		dir:     conf.Dir,
		ssignal: conf.Signal,
		mu:      &sync.Mutex{},
		info: Info{
			Type:   Binary,
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

	// If dies on start, want to make sure this is reset since restart
	// uses the same starting method
	m.userKilled = false

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
	return -1, errors.New("Process.Start.Err=" + perr.Error())
}

// Restart a process.
func (m *LocalManager) Restart(fromFailed bool) (pid int, err error) {
	if !fromFailed {

		m.mu.Lock()
		m.restartCounter = 0
		m.info.Restarts++
		m.userKilled = true
		m.mu.Unlock()

		if m.info.Status == Running || m.info.Status == Stable {
			m.Kill(fromFailed)
			time.Sleep(time.Second * 5)
		}
	}
	return m.Start()
}

// Kill a process
func (m *LocalManager) Kill(withoutRestart bool) {

	m.mu.Lock()
	m.info.Status = Killed
	if withoutRestart {
		m.userKilled = false
	}
	m.mu.Unlock()

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

	listener := make(chan error)
	err := cmd.Start()
	if err != nil {
		m.info.Errors = append(m.info.Errors, err)
		errChan <- err
		pid <- -1
		return
	}

	go func() {
		listener <- cmd.Wait()
	}()

	m.mu.Lock()
	m.info.Status = Running
	m.info.StartTime = time.Now()
	m.info.PID = cmd.Process.Pid
	m.mu.Unlock()

	pid <- cmd.Process.Pid

	select {
	case error := <-listener:
		if err != nil {
			m.info.Errors = append(m.info.Errors, error)
		}

		if m.userKilled {
			return
		}

		if m.gracefulshutdown {
			m.info.Errors = append(m.info.Errors, errors.New("marked for graceful shutdown"))
			return
		}

		if err == nil {
			m.info.Errors = append(
				m.info.Errors,
				errors.New("service shutdown without error; check config; maybe need to turn on blocking"),
			)
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

	if !m.userKilled {

		// Only give up restarting if the process has beeen attempting to restart n times
		// in the last 30 seconds

		if time.Since(m.info.StartTime) > time.Second*30 {
			m.restartCounter = 0
		}

		if m.restartCounter < 3 {
			m.info.Errors = append(m.info.Errors, fmt.Errorf("restart=%d/3; time=%s; last-pid=%d", m.restartCounter+1, time.Now().Format(time.Stamp), m.info.PID))
			m.info.Fails++
			m.restartCounter++
			m.mu.Unlock()
			time.Sleep(time.Second * 5)
			m.Restart(true)
			return
		}
		m.info.Errors = append(m.info.Errors, fmt.Errorf("exceeded restart counter; time=%s; last-pid=%d", time.Now().Format(time.Stamp), m.info.PID))

	}
	m.info.PID = -1
	m.mu.Unlock()
}

// upgrade from Running to Stable if the process.Status is still Running after 30 seconds
func (m *LocalManager) upgrade() {
	time.Sleep(time.Second * 31)
	if m.info.Status == Running {
		if time.Since(m.info.StartTime) > time.Second*30 {
			m.mu.Lock()
			m.info.Status = Stable
			m.mu.Unlock()
		} else {
			m.upgrade()
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
