package metrics

type ConnectedTimeList []float64 // in minutes

func (cl *ConnectedTimeList) AddItem(newItem float64) {
	*cl = append(*cl, newItem)
}

// Get item from the list from index
func (cl *ConnectedTimeList) GetByIndex(idx int) float64 {
	return (*cl)[idx]
}
