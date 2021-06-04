package export

type RBeaconBlockList []int64

func (bb *RBeaconBlockList) AddItem(newItem int64) {
	*bb = append(*bb, newItem)
}

// get item form the list by index
func (bb *RBeaconBlockList) GetByIndex(idx int) int64 {
	return (*bb)[idx]
}
