package metrics

var NodeIdList []string

// add new item to the list
func (pl *NodeIdList) AddItem (newItem string) {
    pl = append(pl, string)
}

// get item form the list by index
func (pl *NodeIdList) GetByItem (idx int) string {
    return pl[idx]
}

// get the array sorted by list of indexes
func (pl *NodeIdList) GetArrayByIndexes(idxs []int) []string {
    var sortedArray []string
    for _, i in range idxs {
        sortedArray = append(sortedArray, pl[i])
    }
    return sortedArray
}

// NOTE: There is no need to sort the peerIds 
