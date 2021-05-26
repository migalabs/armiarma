package export

type RProposerSlashingList []int64

func (bb *RProposerSlashingList) AddItem(newItem int64) {
	*bb = append(*bb, newItem)
}

// get item form the list by index
func (bb *RProposerSlashingList) GetByIndex(idx int) int64 {
	return (*bb)[idx]
}
