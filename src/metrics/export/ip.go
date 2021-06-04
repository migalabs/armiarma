package export

type IpList []string

func (il *IpList) AddItem(newItem string) {
	*il = append(*il, newItem)
}

// Get item from the list from index
func (il *IpList) GetByIndex(idx int) string {
	return (*il)[idx]
}
