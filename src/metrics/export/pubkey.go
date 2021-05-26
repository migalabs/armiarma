package export

type PubKeyList []string

func (pl *PubKeyList) AddItem(newItem string) {
	*pl = append(*pl, newItem)
}

// get item form the list by index
func (pl *PubKeyList) GetByIndex(idx int) string {
	return (*pl)[idx]
}
