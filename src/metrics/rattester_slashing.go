package metrics

type RAttesterSlashingList []int64

func (bb *RAttesterSlashingList) AddItem(newItem int64) {
    *bb = append(*bb, newItem)
}
