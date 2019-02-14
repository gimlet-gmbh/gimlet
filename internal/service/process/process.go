package process

import (
	"time"
)

// Things to think about
// How will untimely process death be notifed

// Status is the enumerated stated of the current status of the service
type Status int

const (
	// Stable ; the process is running and without error for x amunt of time
	Stable Status = 1 + iota

	// Running ; the process is running but not yet stable
	Running

	// Failed ; the process has ran and failed
	Failed

	// Killed ; the process has been killed
	Killed
)

var statuses = [...]string{
	"Stable",
	"Running",
	"Failed",
	"Killed",
}

func (s Status) String() string {
	if Stable <= s && s <= Killed {
		return statuses[s-1]
	}
	return "%!Status()"
}

// Type is the enumerated type of command for the process
type Type int

const (
	// Binary runs the command from the shell
	Binary Type = 1

	// Go runs the command using `go run`
	Go Type = 2

	// Py runs
	// TODO: need to set interpreter
	Py Type = 3

	// Node runs
	// TODO: Learn how to managed node processes
	Node Type = 4
)

// Info is all the reportable information from a process
type Info struct {
	// Type is the type of command that needs to be generated
	// to start the process
	Type Type

	// The status of the process as reported by manager
	Status Status

	// The time that the process was last successfully started
	StartTime time.Time

	// The most recent time that the process has died
	DeathTime time.Time

	// The PID of the process if it has been started, else 0
	PID int

	// The number of times the process has been restarted
	Restarts int

	// The number of times the process has failed
	Fails int

	// The array of errors reported while handling the process
	Errors []error
}

// Manager is the interface to enforce types of process managers
type Manager interface {
	Start() (int, error)
	Kill(withoutRestart bool)
	Restart(fromFailed bool) (int, error)
	GetInfo() Info
	GetErrors() []string
	GetStatus() Status
}
