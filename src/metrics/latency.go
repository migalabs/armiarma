package metrics

type LatencyList []int64

func (ll *LatencyList) AddItem(newItem int64) {// In milliseconds
    *ll = append(*ll, newItem)
}
