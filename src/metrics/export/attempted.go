package export

type AttemptedList []bool

// add new item to the list
func (l *AttemptedList) AddItem(newItem bool) {
	*l = append(*l, newItem)
}

// get item form the list by index
func (l *AttemptedList) GetByIndex(idx int) bool {
	return (*l)[idx]
}

// get the array sorted by list of indexes
func (l AttemptedList) GetArrayByIndexes(idxs []int) []bool {
	var sortedArray []bool
	for _, i := range idxs {
		sortedArray = append(sortedArray, l[i])
	}
	return sortedArray
}
