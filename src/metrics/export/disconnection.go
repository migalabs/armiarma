package export

type DisconnectionList []int64

func (cl *DisconnectionList) AddItem(newItem int64) {
	*cl = append(*cl, newItem)
}

// Get item from the list from index
func (cl *DisconnectionList) GetByIndex(idx int) int64 {
	return (*cl)[idx]
}
