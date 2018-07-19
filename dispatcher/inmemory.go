package dispatcher

import "sync"

type Mem struct {
	Store []string
	sync.Mutex
}

func (c *Mem) Dispatch(j, w string) error {
	c.Lock()
	defer c.Unlock()

	c.Store = append(c.Store, j)
	return nil
}

func NewInMemory() *Mem {
	return &Mem{Store: []string{}}
}
