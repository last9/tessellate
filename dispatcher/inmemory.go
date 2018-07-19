package dispatcher

type mem struct{}

func (c *mem) Dispatch(j, w string) error {
	return nil
}

func NewInMemory() *mem {
	return &mem{}
}
