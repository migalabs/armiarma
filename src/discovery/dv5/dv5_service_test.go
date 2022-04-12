package dv5

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadJSON(t *testing.T) {
	dv5_obj := NewEmptyDiscovery() //NewDefaultConfigData()

	dv5_obj.ImportBootNodeList("./bootnodes_mainnet.json")

	require.Equal(t, dv5_obj.GetBootNodeList()[0].String(), "enr:-KG4QOtcP9X1FbIMOe17QNMKqDxCpm14jcX5tiOE4_TyMrFqbmhPZHK_ZPG2Gxb1GE2xdtodOfx9-cgvNtxnRyHEmC0ghGV0aDKQ9aX9QgAAAAD__________4JpZIJ2NIJpcIQDE8KdiXNlY3AyNTZrMaEDhpehBDbZjM_L9ek699Y7vhUJ-eAdMyQW_Fil522Y0fODdGNwgiMog3VkcIIjKA")

}
