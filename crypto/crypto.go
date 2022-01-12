package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var privateKey *rsa.PrivateKey
var key []byte

func InitCrypto() (err error) {
	key, err = base64.StdEncoding.DecodeString(os.Getenv("aeskey"))
	if err != nil {
		return
	}
	if len(key) != 32 {
		err = errors.New("Wrong aes key length")
		return
	}
	privatePem, err := ioutil.ReadFile("./private.pem")
	if err != nil {
		fmt.Println(err)
		return
	}

	block, _ := pem.Decode(privatePem)
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return
	}
	privateKey = pk.(*rsa.PrivateKey)
	privateKey.Precompute()
	return
}

func Encrypt(data []byte) (baseText string, err error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	baseText = base64.StdEncoding.EncodeToString(ciphertext)
	return
}

func Decrypt(baseText string) (data []byte, err error) {
	ciphertext, err := base64.StdEncoding.DecodeString(baseText)
	if err != nil {
		return
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		err = errors.New("ciphertext size is less than nonceSize")
		return
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	data, err = gcm.Open(nil, nonce, ciphertext, nil)
	return
}

/*
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
*/
