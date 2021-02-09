package metrics

var AddressList []string

func (al *AddressList) NewItem(newItem string) {
    al = append(al, newItem)
}
