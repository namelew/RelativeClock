package server

type Bank struct {
	currentTime uint64
}

func New() *Bank {
	return &Bank{
		currentTime: 0,
	}
}

func (b *Bank) Run() {

}
