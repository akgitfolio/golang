package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/crypto/pbkdf2"
)

const (
	keySize    = 32 // AES-256
	saltSize   = 16
	rsaKeySize = 2048
)

func main() {
	action := flag.String("action", "encrypt", "Action to perform: encrypt or decrypt")
	file := flag.String("file", "", "File to encrypt or decrypt")
	password := flag.String("password", "", "Password for encryption or decryption")
	flag.Parse()

	if *file == "" || *password == "" {
		fmt.Println("File and password must be provided")
		return
	}

	switch *action {
	case "encrypt":
		err := encryptFile(*file, *password)
		if err != nil {
			fmt.Printf("Encryption failed: %v\n", err)
		} else {
			fmt.Println("File encrypted successfully")
		}
	case "decrypt":
		err := decryptFile(*file, *password)
		if err != nil {
			fmt.Printf("Decryption failed: %v\n", err)
		} else {
			fmt.Println("File decrypted successfully")
		}
	default:
		fmt.Println("Unknown action:", *action)
	}
}

func encryptFile(filePath, password string) error {
	plaintext, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	key, salt := deriveKey([]byte(password))
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Prepend the salt to the ciphertext
	result := append(salt, ciphertext...)

	return ioutil.WriteFile(filePath+".enc", result, 0644)
}

func decryptFile(filePath, password string) error {
	ciphertext, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Extract the salt from the ciphertext
	salt := ciphertext[:saltSize]
	ciphertext = ciphertext[saltSize:]

	key := deriveKeyWithSalt([]byte(password), salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath+".dec", plaintext, 0644)
}

func deriveKey(password []byte) ([]byte, []byte) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}
	return pbkdf2.Key(password, salt, 4096, keySize, sha256.New), salt
}

func deriveKeyWithSalt(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 4096, keySize, sha256.New)
}

func generateRSAKeys() (privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, err error) {
	privateKey, err = rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, nil, err
	}
	publicKey = &privateKey.PublicKey
	return privateKey, publicKey, nil
}

func exportRSAPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return privateKeyPEM, nil
}

func exportRSAPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return publicKeyPEM, nil
}

func importRSAPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func importRSAPublicKey(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key is not of type RSA public key")
	}
}

func encryptWithRSA(publicKey *rsa.PublicKey, plaintext []byte) (string, error) {
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext, nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptWithRSA(privateKey *rsa.PrivateKey, ciphertext string) ([]byte, error) {
	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, decodedCiphertext, nil)
}
