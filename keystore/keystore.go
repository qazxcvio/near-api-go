// Package keystore implements an unencrypted file system key store.
package keystore

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcutil/base58"
)

const ed25519Prefix = "ed25519:"

// Ed25519KeyPair is a Ed25519 key pair.
type Ed25519KeyPair struct {
	AccountID      string             `json:"account_id"`
	PublicKey      string             `json:"public_key"`
	PrivateKey     string             `json:"private_key"`
	Ed25519PubKey  ed25519.PublicKey  `json:"-"`
	Ed25519PrivKey ed25519.PrivateKey `json:"-"`
}

// GenerateEd25519KeyPair generates a new Ed25519 key pair for accountID.
func GenerateEd25519KeyPair(accountID string) (*Ed25519KeyPair, error) {
	var (
		kp  Ed25519KeyPair
		err error
	)
	kp.Ed25519PubKey, kp.Ed25519PrivKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	kp.AccountID = accountID
	kp.PublicKey = ed25519Prefix + base58.Encode(kp.Ed25519PubKey)
	kp.PrivateKey = ed25519Prefix + base58.Encode(kp.Ed25519PrivKey)
	return &kp, nil
}

func (kp *Ed25519KeyPair) write(filename string) error {
	data, err := json.Marshal(kp)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}

// Write the Ed25519 key pair to the unencrypted file system key store with
// networkID.
func (kp *Ed25519KeyPair) Write(networkID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filename := filepath.Join(home, ".near-credentials", networkID, kp.AccountID+".json")
	return kp.write(filename)
}

func loadKeyPair(filename, accountID string) (*Ed25519KeyPair, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var kp Ed25519KeyPair
	err = json.Unmarshal(buf, &kp)
	if err != nil {
		return nil, err
	}
	// account ID
	if kp.AccountID != accountID {
		return nil, fmt.Errorf("keystore: parsed account_id '%s' does not match with accountID '%s'",
			kp.AccountID, accountID)
	}
	// public key
	if !strings.HasPrefix(kp.PublicKey, ed25519Prefix) {
		return nil, fmt.Errorf("keystore: parsed public_key '%s' is not an Ed25519 key",
			kp.PublicKey)
	}
	pubKey := base58.Decode(strings.TrimPrefix(kp.PublicKey, ed25519Prefix))
	kp.Ed25519PubKey = ed25519.PublicKey(pubKey)
	// private key
	if !strings.HasPrefix(kp.PrivateKey, ed25519Prefix) {
		return nil, fmt.Errorf("keystore: parsed private_key '%s' is not an Ed25519 key",
			kp.PrivateKey)
	}
	privateKey := base58.Decode(strings.TrimPrefix(kp.PrivateKey, ed25519Prefix))
	kp.Ed25519PrivKey = ed25519.PrivateKey(privateKey)
	// make sure keys match
	if !bytes.Equal(pubKey, kp.Ed25519PrivKey.Public().(ed25519.PublicKey)) {
		return nil, fmt.Errorf("keystore: public_key does not match private_key: %s", filename)
	}
	return &kp, nil
}

// LoadKeyPair reads the Ed25519 key pair for the given networkID and
// accountID form the unencrypted file system key store and returns it.
func LoadKeyPair(networkID, accountID string) (*Ed25519KeyPair, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(home, ".near-credentials", networkID, accountID+".json")
	return loadKeyPair(filename, accountID)
}
