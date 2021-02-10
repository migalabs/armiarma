package metrics

type LatencyList []int64

func (ll *LatencyList) AddItem(newItem int64) { // In milliseconds
	*ll = append(*ll, newItem)
}

// Get item from the list from index
func (ll *LatencyList) GetByIndex(idx int) int64 {
	return (*ll)[idx]
}
