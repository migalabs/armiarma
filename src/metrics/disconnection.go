package metrics

type DisconnectionList []int64

func (cl *DisconnectionList) AddItem(newItem int64) {
    *cl = append(*cl, newItem)
}
