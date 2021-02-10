package metrics

type RVoluntaryExitList []int64

func (bb *RVoluntaryExitList) AddItem(newItem int64) {
    *bb = append(*bb, newItem)
}
