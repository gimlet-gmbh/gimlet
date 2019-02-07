package container

// Container holds data about remote process managers
type Container struct {
	Name    string
	Address string
	ID      string
}

// New returns a new container with name
func New(name string) *Container {
	return &Container{
		Name: name,
	}
}
