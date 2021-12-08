package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

var privateKey *rsa.PrivateKey

func InitCrypto() {
	privatePem, err := ioutil.ReadFile("./private.pem")
	if err != nil {
		fmt.Println(err)
		return
	}

	block, _ := pem.Decode(privatePem)
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err)
		return
	}
	privateKey = pk.(*rsa.PrivateKey)
	privateKey.Precompute()
}

func Encrypt(data string) (txt string, err error) {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha512.New512_256(), rng, &privateKey.PublicKey, []byte(data), label)
	if err != nil {
		return
	}
	txt = base64.URLEncoding.EncodeToString(ciphertext)
	return
}

func Decrypt(data string) (txt string, err error) {
	ct, _ := base64.URLEncoding.DecodeString(data)
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(sha512.New512_256(), rng, privateKey, ct, label)
	if err != nil {
		return
	}
	txt = string(plaintext)
	return
}
