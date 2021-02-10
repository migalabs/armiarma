package metrics

type RBeaconAggregationList []int64

func (bb *RBeaconAggregationList) AddItem(newItem int64) {
    *bb = append(*bb, newItem)
}
