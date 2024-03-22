package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/crypto"
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

func AdaptSecp256k1FromECDSA(ecdsaKey *ecdsa.PrivateKey) (crypto.PrivKey, error) {
	privBytes := gcrypto.FromECDSA(ecdsaKey)
	privKey, err := crypto.UnmarshalSecp256k1PrivateKey(privBytes)
	return privKey, err
}

// Export Private Key to a string
func Secp256k1ToString(inputKey crypto.PrivKey) string {
	keyBytes, _ := inputKey.Raw()
	return hex.EncodeToString(keyBytes)
}

func AdaptECDSAFromSecp256k1(privKey crypto.PrivKey) (*ecdsa.PrivateKey, error) {
	privBytes, err := privKey.Raw()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get bytes from libp2p privkey")
	}
	return gcrypto.ToECDSA(privBytes)
}

// taken from Prysm https://github.com/prysmaticlabs/prysm/blob/616cfd33908df1e479c5dd0980367ede8de82a5d/crypto/ecdsa/utils.go#L38
func ConvertECDSAPubkeyToSecp2561k(pubkey *ecdsa.PublicKey) (crypto.PubKey, error) {
	pubBytes := gcrypto.FromECDSAPub(pubkey)
	secp256k1, err := crypto.UnmarshalSecp256k1PublicKey(pubBytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal libp2p key from geth pubkey bytes")
	}
	return secp256k1, nil
}

func IsLibp2pValidEthereumPrivateKey(privkey crypto.PrivKey) bool {
	secp256privKey, _ := privkey.(*crypto.Secp256k1PrivateKey)
	privBytes, err := secp256privKey.Raw()
	if err != nil {
		return false
	}
	ethCurve := gcrypto.S256()
	ethPrivKey, err := gcrypto.ToECDSA(privBytes)
	if err != nil {
		return false
	}
	return ethCurve.IsOnCurve(ethPrivKey.X, ethPrivKey.Y)
}

func IsLibp2pValidEthereumPublicKey(pubkey crypto.PubKey) bool {
	secp256pubKey, _ := pubkey.(*crypto.Secp256k1PublicKey)
	pubBytes, err := secp256pubKey.Raw()
	if err != nil {
		return false
	}
	ethCurve := gcrypto.S256()
	ecdsaPubKey, err := gcrypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		return true
	}
	return ethCurve.IsOnCurve(ecdsaPubKey.X, ecdsaPubKey.Y)
}

func IsGethValidEthereumPrivateKey(privkey *ecdsa.PrivateKey) bool {
	// create new geth-crypto key to get the curve
	/*
		tempKey, _ := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)
		return privkey.Curve.IsOnCurve(tempKey.X, tempKey.Y)
	*/
	ethCurve := gcrypto.S256()
	return ethCurve.IsOnCurve(privkey.X, privkey.Y)
}

func IsGethValidEthereumPublicKey(pubkey *ecdsa.PublicKey) bool {
	// create new geth-crypto key to get the curve
	ethCurve := gcrypto.S256()
	return ethCurve.IsOnCurve(pubkey.X, pubkey.Y)
}
