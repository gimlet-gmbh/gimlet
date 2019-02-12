package process

import (
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/rpc"
)

// RemoteManager ; remote process manager;  holds data about remote process.Manager
type RemoteManager struct {
	Name     string
	Address  string
	ID       string
	LastInfo Info
}

// NewRemoteManager returns a new container with name
func NewRemoteManager(name, id string) *RemoteManager {
	return &RemoteManager{
		Name: name,
	}
}

// RestartProcess that is being managed by gmbh-container
func (c *RemoteManager) RestartProcess() (string, error) {
	client, ctx, can, err := rpc.GetRemoteRequest(c.Address, time.Second*5)
	if err != nil {
		return "-1", err
	}
	defer can()

	request := &cabal.Action{
		Sender: "gmbh-core",
		Target: c.Name,
		Action: "service.restart",
	}

	response, err := client.RequestRemoteAction(ctx, request)
	if err != nil {
		return "-1", err
	}
	return response.GetStatus(), nil
}

// GetInfo about the remote process
func (c *RemoteManager) GetInfo() Info {
	return c.LastInfo
}
