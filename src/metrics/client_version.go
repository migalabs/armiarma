package metrics

type ClientVersionList []string

// Add new item to the list
func (cv *ClientVersionList) AddItem(newItem string) {
	*cv = append(*cv, newItem)
}

// Get item from the list from index
func (cv *ClientVersionList) GetByIndex(idx int) string {
	return (*cv)[idx]
}

// Add new item to the list
func (cv ClientVersionList) GetArrayByIndexes(idxs []int) []string {
	var sortedArray []string
	for _, i := range idxs {
		sortedArray = append(sortedArray, cv[i])
	}
	return sortedArray
}
