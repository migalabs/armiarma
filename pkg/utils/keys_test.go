package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyConverters(t *testing.T) {
	// Basic Geth privKey
	ogECDSA, err := GenerateECDSAPrivKey()
	require.NoError(t, err)
	ok := IsGethValidEthereumPrivateKey(ogECDSA)
	require.Equal(t, true, ok)

	// Get to Libp2p Secp256k1
	ogLibp2p, err := AdaptSecp256k1FromECDSA(ogECDSA)
	require.NoError(t, err)
	ok = IsLibp2pValidEthereumPrivateKey(ogLibp2p)
	require.Equal(t, true, ok)

	// reverse loop  -> Libp2p to Geth
	rECDSA, err := AdaptECDSAFromSecp256k1(ogLibp2p)
	require.NoError(t, err)
	ok = IsGethValidEthereumPrivateKey(rECDSA)
	require.Equal(t, true, ok)

	// pass Libp2p key to string
	strLibp2p := Secp256k1ToString(ogLibp2p)

	// generate ECDSA from string
	newECDSA, err := ParseECDSAPrivateKey(strLibp2p)
	require.NoError(t, err)
	ok = IsGethValidEthereumPrivateKey(newECDSA)
	require.Equal(t, true, ok)

	// check if ogECDSA is equal as newECDSA
	same := ogECDSA.Equal(newECDSA)
	require.Equal(t, true, same)

	// test the pubkeys
	pubECDSA := newECDSA.PublicKey
	ok = IsGethValidEthereumPublicKey(&pubECDSA)
	require.Equal(t, true, ok)

	pubLibp2p, err := ConvertECDSAPubkeyToSecp2561k(&pubECDSA)
	require.NoError(t, err)
	ok = IsLibp2pValidEthereumPublicKey(pubLibp2p)
	require.Equal(t, true, ok)
}
