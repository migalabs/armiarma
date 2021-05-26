package export

type ErrorList []string

// add new item to the list
func (l *ErrorList) AddItem(newItem string) {
	*l = append(*l, newItem)
}

// get item form the list by index
func (l *ErrorList) GetByIndex(idx int) string {
	return (*l)[idx]
}

// get the array sorted by list of indexes
func (l ErrorList) GetArrayByIndexes(idxs []int) []string {
	var sortedArray []string
	for _, i := range idxs {
		sortedArray = append(sortedArray, l[i])
	}
	return sortedArray
}
