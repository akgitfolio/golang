package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	shell "github.com/ipfs/go-ipfs-api"
)

func EncryptFile(filename string, password string) ([]byte, error) {
	plaintext, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func AddFileToIPFS(data []byte) (string, error) {
	sh := shell.NewShell("localhost:5001")
	hash, err := sh.Add(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	return hash, nil
}

func DecryptFile(data []byte, password string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	if len(data) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return data, nil
}

func main() {
	encryptedData, err := EncryptFile("example.txt", "mysecretpassword")
	if err != nil {
		panic(err)
	}

	hash, err := AddFileToIPFS(encryptedData)
	if err != nil {
		panic(err)
	}

	sh := shell.NewShell("localhost:5001")
	data, err := sh.Cat(hash)
	if err != nil {
		panic(err)
	}

	decryptedData, err := DecryptFile(data, "mysecretpassword")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("decrypted_example.txt", decryptedData, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("File decrypted and saved as decrypted_example.txt")
}
