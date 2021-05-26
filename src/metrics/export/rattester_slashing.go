package export

type RAttesterSlashingList []int64

func (bb *RAttesterSlashingList) AddItem(newItem int64) {
	*bb = append(*bb, newItem)
}

// get item form the list by index
func (bb *RAttesterSlashingList) GetByIndex(idx int) int64 {
	return (*bb)[idx]
}
