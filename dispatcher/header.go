package dispatcher

type Dispatcher interface {
	Dispatch(jobID, workspaceID, layoutID string) error
}

var instance Dispatcher

func Set(d Dispatcher) {
	instance = d
}

func Get() Dispatcher {
	return instance
}
