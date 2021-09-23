package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// Parse a Secp256k1PrivateKey from string, checking if it has the proper curve
func ParsePrivateKey(v string) (*crypto.Secp256k1PrivateKey, error) {
	if strings.HasPrefix(v, "0x") {
		v = v[2:]
	}
	privKeyBytes, err := hex.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key, expected hex string: %v", err)
	}
	var priv crypto.PrivKey
	priv, err = crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key, invalid private key (Secp256k1): %v", err)
	}
	key := (priv).(*crypto.Secp256k1PrivateKey)
	key.Curve = gcrypto.S256()              // Temporary hack, so libp2p Secp256k1 is recognized as geth Secp256k1 in disc v5.1
	if !key.Curve.IsOnCurve(key.X, key.Y) { // TODO: should we be checking this?
		return nil, fmt.Errorf("invalid private key, not on curve")
	}
	return key, nil
}
