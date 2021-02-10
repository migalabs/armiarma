package metrics

type PubKeyList []string

func (pl *PubKeyList) AddItem(newItem string) {
    *pl = append(*pl, newItem)
}


