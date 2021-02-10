package metrics

type IpList []string

func (il *IpList) AddItem(newItem string) {
    *il = append(*il, newItem)
}
