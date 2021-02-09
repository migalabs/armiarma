package metrics

var RVoluntaryExitList []int64

func (bb *RVoluntaryExitList) AddNew(newItem int64) {
    bb = append(bb, newItem)
}
