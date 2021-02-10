package metrics

type TotalMessagesList []int64

func (ll *TotalMessagesList) AddItem(newItem int64) { // In milliseconds
	*ll = append(*ll, newItem)
}

// Get item from the list from index
func (ll *TotalMessagesList) GetByIndex(idx int) int64 {
	return (*ll)[idx]
}
