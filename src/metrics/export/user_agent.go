package export

type UserAgentList []string

func (al *UserAgentList) AddItem(newItem string) {
	*al = append(*al, newItem)
}

func (al *UserAgentList) GetByIndex(idx int) string {
	return (*al)[idx]
}
