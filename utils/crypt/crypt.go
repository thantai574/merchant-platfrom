package crypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

func EncryptSHA256RSA(dataString string, privateKey *rsa.PrivateKey) (res string, err error) {
	h := sha256.New()
	h.Write([]byte(dataString))
	sum := h.Sum(nil)

	sig, _ := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, sum)

	return base64.StdEncoding.EncodeToString(sig), err
}
