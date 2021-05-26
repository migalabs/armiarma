package export

type ConnectedList []bool

// add new item to the list
func (l *ConnectedList) AddItem(newItem bool) {
	*l = append(*l, newItem)
}

// get item form the list by index
func (l *ConnectedList) GetByIndex(idx int) bool {
	return (*l)[idx]
}

// get the array sorted by list of indexes
func (l ConnectedList) GetArrayByIndexes(idxs []int) []bool {
	var sortedArray []bool
	for _, i := range idxs {
		sortedArray = append(sortedArray, l[i])
	}
	return sortedArray
}
