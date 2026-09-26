package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrBadKey = errors.New("bad key")

type KeyPair struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
}

func GenerateKeys() (KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{Private: priv, Public: pub}, nil
}

func LoadOrCreateKeys(privPath, pubPath string) (KeyPair, error) {
	kp, err := LoadKeys(privPath, pubPath)
	if err == nil {
		return kp, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return KeyPair{}, err
	}

	kp, err = GenerateKeys()
	if err != nil {
		return KeyPair{}, err
	}

	if err := SaveKeys(kp, privPath, pubPath); err != nil {
		return KeyPair{}, err
	}

	return kp, nil
}

func LoadKeys(privPath, pubPath string) (KeyPair, error) {
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return KeyPair{}, err
	}
	pubPEM, err := os.ReadFile(pubPath)
	if err != nil {
		return KeyPair{}, err
	}

	priv, err := parsePEM(privPEM, "PRIVATE KEY", func(b []byte) (any, error) { return x509.ParsePKCS8PrivateKey(b) })
	if err != nil {
		return KeyPair{}, fmt.Errorf("%w: %s: %v", ErrBadKey, privPath, err)
	}
	pub, err := parsePEM(pubPEM, "PUBLIC KEY", func(b []byte) (any, error) { return x509.ParsePKIXPublicKey(b) })
	if err != nil {
		return KeyPair{}, fmt.Errorf("%w: %s: %v", ErrBadKey, pubPath, err)
	}

	sk, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return KeyPair{}, fmt.Errorf("%w: %s: not ed25519", ErrBadKey, privPath)
	}

	pk, ok := pub.(ed25519.PublicKey)
	if !ok {
		return KeyPair{}, fmt.Errorf("%w: %s: not ed25519", ErrBadKey, pubPath)
	}

	return KeyPair{Private: sk, Public: pk}, nil
}

func SaveKeys(kp KeyPair, privPath, pubPath string) error {
	privDER, err := x509.MarshalPKCS8PrivateKey(kp.Private)
	if err != nil {
		return err
	}
	pubDER, err := x509.MarshalPKIXPublicKey(kp.Public)
	if err != nil {
		return err
	}
	if err := writePEM(privPath, "PRIVATE KEY", privDER, 0o600); err != nil {
		return err
	}
	return writePEM(pubPath, "PUBLIC KEY", pubDER, 0o644)
}

func parsePEM(data []byte, typ string, parse func([]byte) (any, error)) (any, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != typ {
		return nil, errors.New("no " + typ + " pem block")
	}
	return parse(block.Bytes)
}

func writePEM(path, typ string, der []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), perm)
}
