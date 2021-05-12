package custom

import ()

type ConnectionFailed struct {
	Total int
	TimeOut int
	ResetByPeer int
	DialBackOff int
	DialToSelf int
	Uncertain int
}

func NewConnectionFailed() ConnectionFailed {
	cf := ConnectionFailed {
		Total: 0,
		TimeOut: 0,
		ResetByPeer: 0,
		Uncertain: 0,
		DialBackOff: 0,
		DialToSelf: 0,
	}
	return cf
}

func (cf *ConnectionFailed) SetTotal(v int) {
	cf.Total = v
}

func (cf *ConnectionFailed) SetTimeOut(v int) {
	cf.TimeOut = v
}

func (cf *ConnectionFailed) SetResetByPeer(v int) {
	cf.ResetByPeer = v
}

func (cf *ConnectionFailed) SetDialBackOff(v int) {
	cf.DialBackOff = v
}

func (cf *ConnectionFailed) SetDialToSelf(v int) {
	cf.DialToSelf = v
}

func (cf *ConnectionFailed) SetUncertain(v int) {
	cf.Uncertain = v
}