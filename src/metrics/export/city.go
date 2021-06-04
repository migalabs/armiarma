package export

type CityList []string

func (cl *CityList) AddItem(newItem string) {
	*cl = append(*cl, newItem)
}

func (cl *CityList) GetByIndex(idx int) string {
	return (*cl)[idx]
}
