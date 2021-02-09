package metrics

var IpList []string

func (il *IpList) NewItem(newItem string) {
    il = append(il, newItem)
}
