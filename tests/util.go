package test

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func generateKasRSAKeyPair() ([]byte, []byte, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	pubKey := privKey.PublicKey

	privKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		},
	)

	pk, err := x509.MarshalPKIXPublicKey(&pubKey)
	if err != nil {
		return nil, nil, err
	}

	pubKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pk,
		},
	)

	return privKeyPEM, pubKeyPEM, nil
}

func generateKasECDHKeyPair() ([]byte, []byte, error) {
	privKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	pubKey := privKey.PublicKey()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode the private key in PEM format
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	pk, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, nil, err
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pk,
	})

	return privKeyPEM, pubKeyPEM, nil

}
