package process

import (
	"errors"
	"os/exec"
)

// PyProc fulfills process interface without holding any data. This is used
// for processes that are not in managed mode
type PyProc struct {
	Inf *Info
	Run *Runtime
	Err Perr
}

// NewPyProc returns PyProc
func NewPyProc() *PyProc {
	return &PyProc{}
}

// Start placeholder
func (e *PyProc) Start() (int, error) {
	return -1, errors.New("PyProcess.notImplemented")
}

// Kill placeholder
func (e *PyProc) Kill(withoutRestart bool) {
}

// Restart placeholder
func (e *PyProc) Restart(fromFailed bool) (int, error) {
	return -1, errors.New("PyProcess.notImplemented")
}

// ForkExec placeholder
func (e *PyProc) ForkExec(pid chan int) {
}

func (e *PyProc) getCmd() *exec.Cmd {
	return nil
}

func (e *PyProc) handleFailure() {
}

// GetStatus placeholder
func (e *PyProc) GetStatus() bool {
	return false
}

// GetInfo placeholder
func (e *PyProc) GetInfo() *Info {
	return nil
}

// GetRuntime placeholder
func (e *PyProc) GetRuntime() *Runtime {
	return nil
}

// ReportErrors placeholder
func (e *PyProc) ReportErrors() []string {
	return []string{"PyProcess.notImplemented"}
}
