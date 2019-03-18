package gmbh

import (
	"time"

	"github.com/gmbh-micro/rpc"
)

func (g *Client) connect() {
	g.printer("attempting to connect to coreData")

	if g.state == Connected {
		g.printer("state reported as connected; thread closing")
		return
	}

	reg, status := register()
	for status != nil {
		if status.Error() != "registration.gmbhUnavailable" {
			g.printer("gmbh internal error")
			return
		}

		if g.closed || (g.con != nil && g.con.IsConnected()) {
			return
		}
		g.printer("Could not reach gmbh-core, trying again in 5 seconds")
		time.Sleep(time.Second * 5)
		reg, status = register()

	}

	g.printer("registration details:")
	g.printer("id=" + reg.id + "; address=" + reg.address + "; fingerprint=" + reg.fingerprint)

	if reg.address == "" {
		g.printer("address not received")
		return
	}

	g.mu.Lock()
	g.reg = reg
	g.con = rpc.NewCabalConnection(reg.address, &_server{})
	g.state = Connected

	// add a new channel to communicate to this goroutine
	ph := newPingHelper()
	g.pingHelpers = append(g.pingHelpers, ph)
	g.mu.Unlock()

	err := g.con.Connect()
	if err != nil {
		g.printer("gmbh connection error=(" + err.Error() + ")")
		return
	}
	g.printer("connected; coreAddress=(" + reg.address + ")")

	go g.sendPing(ph)
}
