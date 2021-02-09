package metrics

var CountryList []string

func (cl *CountryList) NewItem(newItem string) {
    cl = append(cl, newItem)
}
