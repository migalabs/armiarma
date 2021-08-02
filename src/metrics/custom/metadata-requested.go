package custom

type MetadataRequested struct {
	/*
	Total      int
	Lighthouse Client
	Teku       Client
	Nimbus     Client
	Prysm      Client
	Lodestar   Client
	Unknown    Client
	*/
}

func NewMetadataRequesed() MetadataRequested {
	mtreq := MetadataRequested{
		/*
		Total:      0,
		Lighthouse: NewClient(),
		Teku:       NewClient(),
		Nimbus:     NewClient(),
		Prysm:      NewClient(),
		Lodestar:   NewClient(),
		Unknown:    NewClient(),
		*/
	}
	return mtreq
}

func (mtreq *MetadataRequested) SetTotal(v int) {
	//mtreq.Total = v
}
