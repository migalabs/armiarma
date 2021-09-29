package utils

import (
	"encoding/hex"
	"strings"

	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/migalabs/armiarma/src/base"
	log "github.com/sirupsen/logrus"
)

// Parse a Secp256k1PrivateKey from string, checking if it has the proper curve
func ParsePrivateKey(v string) (*crypto.Secp256k1PrivateKey, error) {

	localLogger := CreateKeysLogger()

	if strings.HasPrefix(v, "0x") {
		v = v[2:]
	}
	privKeyBytes, err := hex.DecodeString(v)
	if err != nil {
		localLogger.Debugf("cannot parse private key, expected hex string: %v", err)
		return nil, err
	}
	var priv crypto.PrivKey
	priv, err = crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	if err != nil {
		localLogger.Debugf("cannot parse private key, invalid private key (Secp256k1): %v", err)
		return nil, err
	}
	key := (priv).(*crypto.Secp256k1PrivateKey)
	key.Curve = gcrypto.S256()              // Temporary hack, so libp2p Secp256k1 is recognized as geth Secp256k1 in disc v5.1
	if !key.Curve.IsOnCurve(key.X, key.Y) { // TODO: should we be checking this?
		localLogger.Debugf("invalid private key, not on curve")
		return nil, nil
	}
	return key, nil
}

// Export Private Key to a string
func PrivKeyToString(input_key *crypto.Secp256k1PrivateKey) string {

	localLogger := CreateKeysLogger()

	keyBytes, err := input_key.Raw()

	if err != nil {
		localLogger.Debugf("Could not Export Private Key to String")
		return ""
	}

	return hex.EncodeToString(keyBytes)
}

func CreateKeysLogger() log.FieldLogger {
	loggingOpts := base.LogOpts{
		ModName:   "keys",
		Output:    "terminal",
		Formatter: "text",
		Level:     "debug",
	}

	return base.CreateLogger(loggingOpts)
}
