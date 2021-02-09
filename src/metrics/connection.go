package metrics

var ConnetionList []int64

func (cl *ConnectionList) AddNew(newItem int64) {
    cl = append(cl, newItem)
}
