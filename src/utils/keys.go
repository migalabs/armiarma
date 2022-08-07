package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
)

// Parse a Secp256k1PrivateKey from string, checking if it has the proper curve
func ParsePrivateKey(v string) (*crypto.Secp256k1PrivateKey, error) {

	if strings.HasPrefix(v, "0x") {
		v = v[2:]
	}
	privKeyBytes, err := hex.DecodeString(v)
	if err != nil {
		log.Debugf("cannot parse private key, expected hex string: %v", err)
		return nil, err
	}
	var priv crypto.PrivKey
	priv, err = crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	if err != nil {
		log.Debugf("cannot parse private key, invalid private key (Secp256k1): %v", err)
		return nil, err
	}
	key := (priv).(*crypto.Secp256k1PrivateKey)
	// key.Curve = gcrypto.S256()              // Temporary hack, so libp2p Secp256k1 is recognized as geth Secp256k1 in disc v5.1
	// if !key.Curve.IsOnCurve(key.X, key.Y) { // TODO: should we be checking this?
	// 	log.Debugf("invalid private key, not on curve")
	// 	return nil, nil
	// }
	return key, nil
}

// Export Private Key to a string
func PrivKeyToString(input_key *crypto.Secp256k1PrivateKey) string {

	keyBytes, err := input_key.Raw()

	if err != nil {
		// currently Raw() always returns nil, panicking in
		// case that changes in future
		panic(err)
	}

	return hex.EncodeToString(keyBytes)
}

func GeneratePrivKey() *crypto.Secp256k1PrivateKey {
	priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		log.Panicf("failed to generate key: %v", err)
	}
	return priv.(*crypto.Secp256k1PrivateKey)
}

// taken from Prysm https://github.com/prysmaticlabs/prysm/blob/616cfd33908df1e479c5dd0980367ede8de82a5d/crypto/ecdsa/utils.go#L13
func ConvertFromInterfacePrivKey(privkey crypto.PrivKey) (*ecdsa.PrivateKey, error) {
	secpKey := (privkey.(*crypto.Secp256k1PrivateKey))
	rawKey, err := secpKey.Raw()
	if err != nil {
		return nil, err
	}
	privKey := new(ecdsa.PrivateKey)
	k := new(big.Int).SetBytes(rawKey)
	privKey.D = k
	privKey.Curve = gcrypto.S256() // Temporary hack, so libp2p Secp256k1 is recognized as geth Secp256k1 in disc v5.1.
	privKey.X, privKey.Y = gcrypto.S256().ScalarBaseMult(rawKey)
	return privKey, nil
}

// taken from Prysm https://github.com/prysmaticlabs/prysm/blob/616cfd33908df1e479c5dd0980367ede8de82a5d/crypto/ecdsa/utils.go#L38
func ConvertToInterfacePubkey(pubkey *ecdsa.PublicKey) (crypto.PubKey, error) {
	xVal, yVal := new(btcec.FieldVal), new(btcec.FieldVal)
	if xVal.SetByteSlice(pubkey.X.Bytes()) {
		return nil, errors.Errorf("X value overflows")
	}
	if yVal.SetByteSlice(pubkey.Y.Bytes()) {
		return nil, errors.Errorf("Y value overflows")
	}
	newKey := crypto.PubKey((*crypto.Secp256k1PublicKey)(btcec.NewPublicKey(xVal, yVal)))
	// Zero out temporary values.
	xVal.Zero()
	yVal.Zero()
	return newKey, nil
}
