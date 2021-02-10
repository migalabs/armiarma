package metrics

type RBeaconBlockList []int64

func (bb *RBeaconBlockList) AddItem(newItem int64) {
    *bb = append(*bb, newItem)
}
