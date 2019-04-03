package gmbh

import (
	"time"

	"github.com/gmbh-micro/rpc"
)

func (g *Client) connect() {
	print("attempting to connect to coreData")

	if g.state == Connected {
		print("state reported as connected; thread closing")
		return
	}

	reg, status := register()
	for status != nil {
		if status.Error() != "registration.gmbhUnavailable" {
			print("gmbh internal error")
			return
		}

		if g.closed || (g.con != nil && g.con.IsConnected()) {
			return
		}
		print("Could not reach gmbh-core, trying again in 5 seconds")
		time.Sleep(time.Second * 5)
		reg, status = register()

	}

	print("registration details:")
	print("id=" + reg.id + "; address=" + reg.address + "; fingerprint=" + reg.fingerprint)

	if reg.address == "" {
		print("address not received")
		return
	}

	g.mu.Lock()
	g.reg = reg
	g.con = rpc.NewCabalConnection(reg.address, &_server{})
	g.state = Connected

	g.mu.Unlock()

	err := g.con.Connect()
	if err != nil {
		print("gmbh connection error=(" + err.Error() + ")")
		return
	}
	print("connected; coreAddress=(" + reg.address + ")")

}
