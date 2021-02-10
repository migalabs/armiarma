package metrics

type CountryList []string

func (cl *CountryList) AddItem(newItem string) {
    *cl = append(*cl, newItem)
}
