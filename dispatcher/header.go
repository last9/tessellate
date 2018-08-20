package dispatcher

import "github.com/tsocial/tessellate/storage/types"

type Dispatcher interface {
	Dispatch(workspaceID string, job *types.Job) error
}

var instance Dispatcher

func Set(d Dispatcher) {
	instance = d
}

func Get() Dispatcher {
	return instance
}
