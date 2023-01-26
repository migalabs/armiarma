package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"

	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
)

// Generate PrivateKey valid for Ethereum Consensus Layer
func GenerateECDSAPrivKey() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate key")
	}
	return key, nil
}

// Parse a Secp256k1PrivateKey from string (Libp2p), checking if it has the proper curve
func ParseECDSAPrivateKey(strKey string) (*ecdsa.PrivateKey, error) {
	return gcrypto.HexToECDSA(strKey)
}

func AdaptSecp256k1FromECDSA(ecdsaKey *ecdsa.PrivateKey) (*crypto.Secp256k1PrivateKey, error) {
	secpKey := (*crypto.Secp256k1PrivateKey)(ecdsaKey)
	return secpKey, nil
}

// Export Private Key to a string
func Secp256k1ToString(inputKey *crypto.Secp256k1PrivateKey) string {
	keyBytes, _ := inputKey.Raw()
	return hex.EncodeToString(keyBytes)
}

func AdaptECDSAFromSecp256k1(privKey *crypto.Secp256k1PrivateKey) (*ecdsa.PrivateKey, error) {
	privBytes, err := privKey.Raw()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get bytes from libp2p privkey")
	}
	return gcrypto.ToECDSA(privBytes)
}

// taken from Prysm https://github.com/prysmaticlabs/prysm/blob/616cfd33908df1e479c5dd0980367ede8de82a5d/crypto/ecdsa/utils.go#L38
func ConvertECDSAPubkeyToSecp2561k(pubkey *ecdsa.PublicKey) (*crypto.Secp256k1PublicKey, error) {
	pubBytes := gcrypto.FromECDSAPub(pubkey)
	secp256k1, err := crypto.UnmarshalSecp256k1PublicKey(pubBytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal libp2p key from geth pubkey bytes")
	}
	return secp256k1.(*crypto.Secp256k1PublicKey), nil
}

func IsLibp2pValidEthereumPrivateKey(privkey *crypto.Secp256k1PrivateKey) bool {
	tempKey, _ := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)
	return privkey.IsOnCurve(tempKey.X, tempKey.Y)
}

func IsLibp2pValidEthereumPublicKey(pubkey *crypto.Secp256k1PublicKey) bool {
	temPubkey, _ := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)
	return pubkey.Curve.IsOnCurve(temPubkey.X, temPubkey.Y)
}

func IsGethValidEthereumPrivateKey(privkey *ecdsa.PrivateKey) bool {
	// create new geth-crypto key to get the curve
	tempKey, _ := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)
	return privkey.Curve.IsOnCurve(tempKey.X, tempKey.Y)
}

func IsGethValidEthereumPublicKey(pubkey *ecdsa.PublicKey) bool {
	// create new geth-crypto key to get the curve
	tempKey, _ := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)
	return pubkey.Curve.IsOnCurve(tempKey.X, tempKey.Y)
}
