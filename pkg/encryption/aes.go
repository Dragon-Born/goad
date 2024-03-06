package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding removes the PKCS7 padding from a text
func PKCS7UnPadding(plaintext []byte) ([]byte, error) {
	length := len(plaintext)
	if length == 0 {
		return nil, nil // or error for empty input
	}
	padding := int(plaintext[length-1])
	return plaintext[:length-padding], nil
}

func GenerateAESKey(passphrase string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(passphrase))
	return hasher.Sum(nil)
}

func EncryptHex(text string, passphrase string) (string, error) {
	key := GenerateAESKey(string(GenerateAESKey(passphrase)) + passphrase)
	plaintext, err := hex.DecodeString(text)
	if err != nil {
		return "", err
	}
	plaintext = PKCS7Padding(plaintext, aes.BlockSize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return hex.EncodeToString(ciphertext), nil
}

func DecryptHex(text string, passphrase string) (string, error) {
	key := GenerateAESKey(string(GenerateAESKey(passphrase)) + passphrase)
	ciphertext := []byte(text)
	ciphertext, err := hex.DecodeString(text)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", err // or a more specific error
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS#7 padding
	plaintext, err := PKCS7UnPadding(ciphertext)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(plaintext), nil
}
