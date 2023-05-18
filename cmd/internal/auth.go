package internal

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Authenticator interface {
	Authenticate(username, signature string, msg []byte) (bool, error)
}

type InMemoryAuthenticator struct {
	users map[string]*rsa.PublicKey
}

func NewInMemoryAuthenticator() *InMemoryAuthenticator {
	// Here we hardcode users and their keys
	users, err := readFromSshDir()
	if err != nil {
		log.Panicln("Error creating in memory authenticator")
	}

	// Note: For the sake of example, let's suppose we have the public keys.
	//       In real situation, you'd want to load these keys from a secure source
	user1PublicKey := &rsa.PublicKey{ /* user 1's public key data */ }
	user2PublicKey := &rsa.PublicKey{ /* user 2's public key data */ }

	users["user1"] = user1PublicKey
	users["user2"] = user2PublicKey

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
	signatureBytes := []byte(signature)

	// Verify the signature
	err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashedMsg, signatureBytes)
	if err != nil {
		return false, nil // The authentication failed, but this is not an 'error' per se
	}

	return true, nil
}

func readFromSshDir() (map[string]*rsa.PublicKey, error) {
	users := make(map[string]*rsa.PublicKey)
	sshDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sshDir = filepath.Join(sshDir, ".ssh")

	files, err := ioutil.ReadDir(sshDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".pub" {
			content, err := ioutil.ReadFile(filepath.Join(sshDir, file.Name()))
			if err != nil {
				return nil, err
			}

			block, _ := pem.Decode(content)
			if block == nil || block.Type != "PUBLIC KEY" {
				continue
			}

			publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}

			users[file.Name()[:len(file.Name())-4]] = publicKey
		}
	}
	return users, nil
}
