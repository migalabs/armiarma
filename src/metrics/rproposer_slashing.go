package metrics

type RProposerSlashingList []int64

func (bb *RProposerSlashingList) AddItem(newItem int64) {
    *bb = append(*bb, newItem)
}
