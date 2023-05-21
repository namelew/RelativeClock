package client

type Teller struct {
	currentTime uint64
}

func New() *Teller {
	return &Teller{
		currentTime: 0,
	}
}

func (t *Teller) Run() {
}
