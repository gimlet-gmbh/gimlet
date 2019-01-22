package gmbh

import "github.com/gimlet-gmbh/gimlet/ipc"

var c *Cabal

type config struct {
	isServer bool
	isClient bool
	address  string
}

// Cabal - singleton for cabal coms
type Cabal struct {
	name     string
	isServer bool
	isClient bool
	address  string

	registeredFunctions map[string]func(req ipc.Request, resp *ipc.Responder)
}
