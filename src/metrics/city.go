package metrics

type CityList []string

func (cl *CityList) AddItem(newItem string) {
    *cl = append(*cl, newItem)
}
