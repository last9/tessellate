package dispatcher

import (
	"sync"

	"github.com/tsocial/tessellate/storage/types"
)

type Mem struct {
	Store []string
	sync.Mutex
}

func (c *Mem) Dispatch(w string, j *types.Job) error {
	c.Lock()
	defer c.Unlock()

	c.Store = append(c.Store, j.Id)
	return nil
}

func NewInMemory() *Mem {
	return &Mem{Store: []string{}}
}
