package message

type TopicMessage interface {
	DataParser()
	GetContent()
}

type BeaconBlockMessage struct {
}

type AttestationBlockMessage struct {
}

type ProposerSlashingMessage struct {
}

type AttesterSlashingMessage struct {
}

type VoluntaryExitMessage struct {
}
