package export

type ConnectionList []int64

func (cl *ConnectionList) AddItem(newItem int64) {
	*cl = append(*cl, newItem)
}

// Get item from the list from index
func (cl *ConnectionList) GetByIndex(idx int) int64 {
	return (*cl)[idx]
}
