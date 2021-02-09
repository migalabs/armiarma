package metrics

var RBeaconAggregationList []int64

func (bb *RBeaconAggregationList) AddNew(newItem int64) {
    bb = append(bb, newItem)
}
