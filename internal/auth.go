package internal

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

type Authenticator interface {
	Authenticate(username, signature string, msg []byte) (bool, error)
}

type InMemoryAuthenticator struct {
	users map[string]*rsa.PublicKey
}

func NewInMemoryAuthenticator(authDir string) *InMemoryAuthenticator {
	// Here we hardcode users and their keys
	users, err := readUsersFromDir(authDir)
	if err != nil {
		fmt.Println(err)
		log.Panicln("Error creating in memory authenticator")
	}
	return &InMemoryAuthenticator{users: users}
}

func (a *InMemoryAuthenticator) Authenticate(username, signature string, msg []byte) (bool, error) {
	pubKey, ok := a.users[username]
	if !ok {
		return false, errors.New("unknown user")
	}

	// Generate a cryptographic hash of the message
	hash := sha256.New()
	hash.Write(msg)
	hashedMsg := hash.Sum(nil)

	// Convert the signature from a string back to a byte slice for verification
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err // Handle this error properly
	}

	// Verify the signature
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashedMsg, signatureBytes)
	if err != nil {
		fmt.Println("username:", username)
		fmt.Println("hashed message:", hashedMsg)
		fmt.Println("decoded signature:", signatureBytes)
		return false, nil // The authentication failed, but this is not an 'error' per se
	}

	return true, nil
}

func readUsersFromDir(dir string) (map[string]*rsa.PublicKey, error) {
	users := make(map[string]*rsa.PublicKey)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".pub" {
			content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}

			block, _ := pem.Decode(content)
			if block == nil || block.Type != "PUBLIC KEY" {
				log.Printf("Could not decode %s", file.Name())
				continue
			}

			publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}

			rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
			if !ok {
				return nil, errors.New("invalid RSA public key")
			}

			users[file.Name()[:len(file.Name())-4]] = rsaPublicKey
		}
	}
	return users, nil
}
