package di

type DI struct {
}

func New() *DI {
	return &DI{}
}

func (d *DI) mustExit(err error) {
	panic(err)
}
