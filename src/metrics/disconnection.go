package metrics

var DisconnetionList []int64

func (cl *DisconnectionList) AddNew(newItem int64) {
    cl = append(cl, newItem)
}
