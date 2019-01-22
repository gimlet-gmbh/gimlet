package grouter

import "strconv"

// Router is used for handling address assignment during new service
// discovery
type Router struct {
	BaseAddress    string
	NextPortNumber int
	Localhost      bool
}

func (r *Router) GetNextAddress() string {
	addr := r.BaseAddress + strconv.Itoa(r.NextPortNumber)
	r.chooseNextPort()
	return addr
}

func (r *Router) chooseNextPort() {
	r.NextPortNumber += 10
}
