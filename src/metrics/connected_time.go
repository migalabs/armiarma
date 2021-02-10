package metrics

type ConnectedTimeList []int64 // in minutes

func (cl *ConnectedTimeList) AddItem(newItem int64) {
	*cl = append(*cl, newItem)
}

// Get item from the list from index
func (cl *ConnectedTimeList) GetByIndex(idx int) int64 {
	return (*cl)[idx]
}
