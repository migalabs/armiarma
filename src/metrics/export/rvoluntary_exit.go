package export

type RVoluntaryExitList []int64

func (bb *RVoluntaryExitList) AddItem(newItem int64) {
	*bb = append(*bb, newItem)
}

// get item form the list by index
func (bb *RVoluntaryExitList) GetByIndex(idx int) int64 {
	return (*bb)[idx]
}
