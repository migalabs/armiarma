package metrics

var RBeaconBlockList []int64

func (bb *RBeaconBlockList) AddNew(newItem int64) {
    bb = append(bb, newItem)
}
