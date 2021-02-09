package metrics

var RAttesterSlashingList []int64

func (bb *RAttesterSlashingList) AddNew(newItem int64) {
    bb = append(bb, newItem)
}
