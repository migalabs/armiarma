package export

type AddressList []string

func (al *AddressList) AddItem(newItem string) {
	*al = append(*al, newItem)
}

func (al *AddressList) GetByIndex(idx int) string {
	return (*al)[idx]
}
