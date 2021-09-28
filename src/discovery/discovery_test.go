package discovery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadJSON(t *testing.T) {
	dv5_obj := NewEmptyDiscovery() //NewDefaultConfigData()

	dv5_obj.ImportBootNodeList("bootnodes_mainnet.json")

	fmt.Println(dv5_obj.GetBootNodeList()[0].String())

	require.Equal(t, dv5_obj.GetBootNodeList()[0].String(), "enr:-Ku4QImhMc1z8yCiNJ1TyUxdcfNucje3BGwEHzodEZUan8PherEo4sF7pPHPSIB1NNuSg5fZy7qFsjmUKs2ea1Whi0EBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQOVphkDqal4QzPMksc5wnpuC3gvSC8AfbFOnZY_On34wIN1ZHCCIyg")

}
