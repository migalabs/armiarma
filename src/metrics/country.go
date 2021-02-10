package metrics

type CountryList []string

func (cl *CountryList) AddItem(newItem string) {
	*cl = append(*cl, newItem)
}

// Get item from the list from index
func (cl *CountryList) GetByIndex(idx int) string {
	return (*cl)[idx]
}
