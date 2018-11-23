package dispatcher

import (
	"sync"

	"github.com/tsocial/tessellate/storage/types"
)

type Mem struct {
	Store []string
	sync.Mutex
}

func (c *Mem) Dispatch(w string, j *types.Job) (string, error) {
	c.Lock()
	defer c.Unlock()

	c.Store = append(c.Store, j.Id)
	return j.Id, nil
}

func NewInMemory() *Mem {
	return &Mem{Store: []string{}}
}
