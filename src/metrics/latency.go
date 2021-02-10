package metrics

type LatencyList []float64

func (ll *LatencyList) AddItem(newItem float64) { // In milliseconds
	*ll = append(*ll, newItem)
}

// Get item from the list from index
func (ll *LatencyList) GetByIndex(idx int) float64 {
	return (*ll)[idx]
}
