package gmbh

import (
	"strconv"
	"sync"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
	"google.golang.org/grpc/metadata"
)

// sendPing is meant to run in its own thread. It will continue to call itself or
// return and changed the state of the connection if there is a failure reaching
// the control server that is ran by gmbhCore
func (g *Client) sendPing(ph *pingHelper) {

	// Loop forever
	for {

		time.Sleep(g.PongTime)
		// g.printer("-> ping")

		select {
		case _ = <-ph.pingChan: // case in which this channel is no longer needed
			g.printer("received chan message in sendPing")
			close(ph.pingChan)
			ph.mu.Lock()
			ph.received = true
			ph.mu.Unlock()
			return
		default: // default operation, wait and send a ping

			if !g.con.IsConnected() {
				return
			}

			client, ctx, can, err := rpc.GetCabalRequest(config.DefaultSystemCore.Address, time.Second*30)
			if err != nil {
				g.printer(err.Error())
			}

			if g.reg == nil {
				g.printer("invalid reg for ping")
				return
			}

			ctx = metadata.AppendToOutgoingContext(
				ctx,
				"sender", g.opts.service.Name,
				"target", "procm",
				"fingerprint", g.reg.fingerprint,
			)

			pong, err := client.Alive(ctx, &intrigue.Ping{
				Time: time.Now().Format(time.Stamp),
			})
			if err != nil {
				g.failed()
				return
			}
			if pong.GetStatus() == "core.verified" {
				can()
				// g.printer("<- pong")
			} else {
				g.printer("<- pong err=" + pong.GetError())
				g.failed()
				return
			}
		}
	}
}

type pingHelper struct {
	pingChan  chan bool
	contacted bool
	received  bool
	mu        *sync.Mutex
}

func newPingHelper() *pingHelper {
	return &pingHelper{
		pingChan: make(chan bool, 1),
		mu:       &sync.Mutex{},
	}
}

func update(phs []*pingHelper) []*pingHelper {
	n := []*pingHelper{}
	c := 0
	for _, p := range phs {
		if p.contacted && p.received {
			n = append(n, p)
		} else {
			c++
		}
	}
	g.printer("removed " + strconv.Itoa(len(phs)-c) + "/" + strconv.Itoa(len(phs)) + " channels")
	return n
}
