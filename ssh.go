package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Most code for this taken from: https://gist.github.com/goliatone/e9c13e5f046e34cef6e150d06f20a34c

// Generate a new SSH key pair
func generateSSHKeyPair() ([]byte, []byte, error) {
	bitSize := 2048
	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil, err
	}

	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil, err
	}

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	// Store public and private key to lcoal file system
	err = os.WriteFile("id_rsa", privateKeyBytes, 0600)
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil, err
	}

	err = os.WriteFile("id_rsa.pub", publicKeyBytes, 0644)
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil, err
	}

	return privateKeyBytes, publicKeyBytes, nil
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Println("Public key generated")
	return pubKeyBytes, nil
}

// Create a new SSH client
func createSSHClient(privateKey []byte, ipAddress string) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var client *ssh.Client
	for {
		client, err = ssh.Dial("tcp", ipAddress+":22", config)
		if err != nil {
			fmt.Println("Waiting for SSH connection...")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	return client, nil
}

// Run a command on the remote server
func runCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(command)
}

// Run multiple commands on SSH session
func runCommands(client *ssh.Client, commands []string) error {

	fmt.Println("Running commands...")
	for _, command := range commands {
		fmt.Println("Running command: " + command)
		session, err := client.NewSession()
		if err != nil {
			return err
		}
		defer session.Close()

		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
		err = session.Run(command)
		if err != nil {
			return err
		}
	}

	return nil
}
