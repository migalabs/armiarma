package export

type CountryCodeList []string

func (cl *CountryCodeList) AddItem(newItem string) {
	*cl = append(*cl, newItem)
}

// Get item from the list from index
func (cl *CountryCodeList) GetByIndex(idx int) string {
	return (*cl)[idx]
}
