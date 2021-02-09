package metrics

var ConnectedTimeList []int64 // in minutes

func (cl *ConnectedTimeList) AddNew(newItem int64) {
    cl = append(cl, newItem)
}
