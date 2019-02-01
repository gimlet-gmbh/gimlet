package process

import (
	"errors"
	"os/exec"
)

// NodeProc fulfills process interface without holding any data. This is used
// for processes that are not in managed mode
type NodeProc struct {
	Inf *Info
	Run *Runtime
	Err Perr
}

// NewNodeProc returns NodeProc
func NewNodeProc() *NodeProc {
	return &NodeProc{}
}

// Start placeholder
func (e *NodeProc) Start() (int, error) {
	return -1, errors.New("NodeProcess.notImplemented")
}

// Kill placeholder
func (e *NodeProc) Kill(withoutRestart bool) {
}

// Restart placeholder
func (e *NodeProc) Restart(fromFailed bool) (int, error) {
	return -1, errors.New("NodeProcess.notImplemented")
}

// ForkExec placeholder
func (e *NodeProc) ForkExec(pid chan int) {
}

func (e *NodeProc) getCmd() *exec.Cmd {
	return nil
}

func (e *NodeProc) handleFailure() {
}

// GetStatus placeholder
func (e *NodeProc) GetStatus() bool {
	return false
}

// GetInfo placeholder
func (e *NodeProc) GetInfo() *Info {
	return nil
}

// GetRuntime placeholder
func (e *NodeProc) GetRuntime() *Runtime {
	return nil
}

// ReportErrors placeholder
func (e *NodeProc) ReportErrors() []string {
	return []string{"NodeProcess.notImplemented"}
}
