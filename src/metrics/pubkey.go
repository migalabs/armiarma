package metrics

var PubKeyList []string

func (pl *PubKeyList) AddItem(newItem string) {
    pl = append(pl, newItem)
}


