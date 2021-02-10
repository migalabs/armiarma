package metrics

type RBeaconAggregationList []int64

func (bb *RBeaconAggregationList) AddItem(newItem int64) {
	*bb = append(*bb, newItem)
}

// get item form the list by index
func (bb *RBeaconAggregationList) GetByIndex(idx int) int64 {
	return (*bb)[idx]
}
