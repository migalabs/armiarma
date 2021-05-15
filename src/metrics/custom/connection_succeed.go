package custom

import ()

type ConnectionSucceed struct {
	Total int
	Lighthouse Client
	Teku Client
	Nimbus Client
	Prysm Client
	Lodestar Client
	Unknown Client
}

func NewConnectionSucceed() ConnectionSucceed {
	cs := ConnectionSucceed {
		Total: 0,
		Lighthouse: NewClient(),
		Teku: NewClient(),
		Nimbus: NewClient(),
		Prysm: NewClient(),
		Lodestar: NewClient(),
		Unknown: NewClient(),
	}
	return cs
}

func (cs *ConnectionSucceed) SetTotal(v int) {
	cs.Total = v
}