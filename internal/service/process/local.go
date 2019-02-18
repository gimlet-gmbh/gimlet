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
	name             string
	args             []string
	env              []string
	path             string
	dir              string
	userKilled       bool
	userRestarted    bool
	restartCounter   int
	gracefulshutdown bool
	ssignal          syscall.Signal
	mu               *sync.Mutex
	info             Info
}

// NewLocalBinaryManager ; as in new process manager to monitor a binary forked from the shell
func NewLocalBinaryManager(name string, path string, dir string, args []string, env []string, shutdownSignal syscall.Signal) *LocalManager {
	return &LocalManager{
		name:    name,
		args:    args,
		env:     env,
		path:    path,
		dir:     dir,
		ssignal: shutdownSignal,
		mu:      &sync.Mutex{},
		info: Info{
			Type:   Binary,
			Errors: make([]error, 0),
		},
	}
}

// NewLocalGoManager ; as in new process manager to monitor bulding from go source code
func NewLocalGoManager(name string, path string, dir string, args []string, env []string) *LocalManager {
	return &LocalManager{
		name: name,
		args: args,
		env:  env,
		path: path,
		dir:  dir,
		mu:   &sync.Mutex{},
		info: Info{
			Type:   Go,
			Errors: make([]error, 0),
		},
	}
}

// Start a process if possible
func (m *LocalManager) Start() (PID int, err error) {

	// If dies on start, want to make sure this is reset since restart
	// uses the same starting method
	m.userKilled = false

	getPidChan := make(chan int, 1)
	go m.forkExec(getPidChan)
	pid := <-getPidChan
	if pid != -1 {
		m.mu.Lock()
		m.info.Status = Running
		m.mu.Unlock()
		go m.upgrade()
		return pid, nil
	}
	return -1, errors.New("process.Manager.Start.failure")
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

func (m *LocalManager) forkExec(pid chan int) {

	cmd := m.getCmd()

	// file, err := m.createLogFile(m.dir+defaults.SERVICE_LOG_PATH, defaults.SERVICE_LOG_FILE)
	// if err != nil {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// } else {
	// 	cmd.Stdout = file
	// 	cmd.Stderr = file
	// }

	listener := make(chan error)
	err := cmd.Start()
	if err != nil {
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

		m.handleFailure()
	}

}

func (m *LocalManager) getCmd() *exec.Cmd {
	if m.info.Type == Binary {
		var cmd *exec.Cmd
		cmd = exec.Command(m.path, m.args...)
		cmd.Env = m.env
		fmt.Println("using " + m.dir + " as dir")
		cmd.Dir = m.dir
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

func (m *LocalManager) createLogFile(path, filename string) (*os.File, error) {

	m.checkDir(path)

	stdOut, err := os.OpenFile(path+filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return nil, errors.New("could not create log file")
	}

	return stdOut, nil
}

func (m *LocalManager) checkDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// GracefulShutdown blocks restart on failure
func (m *LocalManager) GracefulShutdown() {
	m.mu.Lock()
	m.gracefulshutdown = true
	m.mu.Unlock()
}
