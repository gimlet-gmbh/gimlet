package procm

import (
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/rpc"
)

// Manager ; process manager;  holds data about remote process managers
type Manager struct {
	Name    string
	Address string
	ID      string
}

// New returns a new container with name
func New(name string) *Manager {
	return &Manager{
		Name: name,
	}
}

// RestartProcess that is being managed by gmbh-container
func (c *Manager) RestartProcess() {
	client, ctx, can, err := rpc.GetRemoteRequest(c.Address, time.Second)
	if err != nil {
		return
	}
	defer can()

	request := &cabal.Action{
		Sender:  "gmbh-core",
		Target:  c.Name,
		Message: "service.restart",
	}

	client.RequestRemoteAction(ctx, request)
}
