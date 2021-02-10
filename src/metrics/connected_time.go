package metrics

type ConnectedTimeList []int64 // in minutes

func (cl *ConnectedTimeList) AddItem(newItem int64) {
    *cl = append(*cl, newItem)
}
