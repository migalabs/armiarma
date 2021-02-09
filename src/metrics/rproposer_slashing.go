package metrics

var RProposerSlashingList []int64

func (bb *RProposerSlashingList) AddNew(newItem int64) {
    bb = append(bb, newItem)
}
