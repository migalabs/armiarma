package metrics

type AddressList []string

func (al AddressList) AddItem(newItem string) {
    al = append(al, newItem)
}
