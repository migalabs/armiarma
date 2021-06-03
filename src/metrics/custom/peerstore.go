package custom

type PeerStore struct {
	Total                 int
	Port13000             int
	Port9000              int
	PortDiff              int
	NoAttemptedConnection int
	ConnectionFailed      ConnectionFailed
	ConnectionSucceed     ConnectionSucceed
	MetadataRequested     MetadataRequested
}

func NewPeerStore() PeerStore {
	ps := PeerStore{
		Total:                 0,
		Port13000:             0,
		Port9000:              0,
		PortDiff:              0,
		NoAttemptedConnection: 0,
		ConnectionFailed:      NewConnectionFailed(),
		ConnectionSucceed:     NewConnectionSucceed(),
		MetadataRequested:     NewMetadataRequesed(),
	}
	return ps
}

func (ps *PeerStore) SetTotal(t int) {
	ps.Total = t
}

func (ps *PeerStore) SetPort13000(t int) {
	ps.Port13000 = t
}

func (ps *PeerStore) SetPort9000(t int) {
	ps.Port9000 = t
}

func (ps *PeerStore) SetPortDiff(t int) {
	ps.PortDiff = t
}

func (ps *PeerStore) SetNotAttempted(t int) {
	ps.NoAttemptedConnection = t
}
