package metrics

type ClientTypeList []string

// Add new item to the list
func (ct *ClientTypeList) AddItem(newItem string) {
	*ct = append(*ct, newItem)
}

// Get item from the list from index
func (ct *ClientTypeList) GetByIndex(idx int) string {
	return (*ct)[idx]
}

// Add new item to the list
func (ct ClientTypeList) GetArrayByIndexes(idxs []int) []string {
	var sortedArray []string
	for _, i := range idxs {
		sortedArray = append(sortedArray, ct[i])
	}
	return sortedArray
}
