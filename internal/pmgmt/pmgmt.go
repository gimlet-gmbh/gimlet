package pmgmt

/**
 * pmgmt.go
 * Abe Dick
 * January 2019
 */

// NewGoProcess returns a new golang process
func NewGoProcess(name, path string) *Process {
	p := Process{
		// id: getProcID(),
		Controller: &GoProcess{},
		Info: pInfo{
			name: name,
			args: []string{},
			path: path,
			// build: build,
		},
		Runtime: pRuntime{
			running:     false,
			userKilled:  false,
			Pid:         -1,
			numRestarts: 0,
		},
		Errs: pError{
			errors: []error{},
		},
	}

	return &p
}
