package metrics

type ConnectionList []int64

func (cl *ConnectionList) AddItem(newItem int64) {
    *cl = append(*cl, newItem)
}
