package address

import (
	"fmt"
	"strconv"
	"sync"
)

// Handler ; as in address handler. Manages the assignemnt of addresses
type Handler struct {
	host        string
	portHigh    int
	portLow     int
	currentPort int
	table       map[string]string
	usedPorts   map[int]bool
	mu          *sync.Mutex
}

// NewHandler returns a new address handler
func NewHandler(host string, portLow, portHigh int) *Handler {
	return &Handler{
		host:        host,
		portLow:     portLow,
		currentPort: portLow,
		portHigh:    portHigh,
		table:       make(map[string]string),
		usedPorts:   make(map[int]bool),
		mu:          &sync.Mutex{},
	}
}

// NextAddress assignes the next address of the handler
func (h *Handler) NextAddress() (string, error) {
	next, err := h.nextPort()
	if err != nil {
		return "", err
	}
	return h.host + ":" + strconv.Itoa(next), nil
}

// nextPort returns the next port number and increments the current port
func (h *Handler) nextPort() (int, error) {
	if h.currentPort+2 < h.portHigh {
		port := h.currentPort + 2
		h.currentPort += 2
		return port, nil
	}
	return -1, fmt.Errorf("out of port range")
}
