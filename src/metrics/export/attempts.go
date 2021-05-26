package export

type AttemptsList []int

// add new item to the list
func (l *AttemptsList) AddItem(newItem int) {
	*l = append(*l, newItem)
}

// get item form the list by index
func (l *AttemptsList) GetByIndex(idx int) int {
	return (*l)[idx]
}

// get the array sorted by list of indexes
func (l AttemptsList) GetArrayByIndexes(idxs []int) []int {
	var sortedArray []int
	for _, i := range idxs {
		sortedArray = append(sortedArray, l[i])
	}
	return sortedArray
}
