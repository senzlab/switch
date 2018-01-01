package main

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/asn1"
    "encoding/pem"
    "fmt"
    "os"
)

var keySize = 1024

func initSwitchKey() {
    // check keys exists
    if _, err := os.Stat(config.dotKeys); err == nil {
        println("keys exists")
        return
    }

    // create keys directory
    err := os.Mkdir(config.dotKeys, 0700)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }

    // generate key pair
    // save private key
    // save public key
    keyPair := initKey()
    savePrivateKey(config.idRsa, keyPair)
    savePublicKey(config.idRsaPub, keyPair)
}

func initKeyPair() *rsa.PrivateKey {
    // generate key pair
    keyPair, err := rsa.GenerateKey(rand.Reader, keySize)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }

    // validate key
	err = keyPair.Validate()
	if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
	}

    return keyPair
}

func savePrivateKey(fileName string, keyPair *rsa.PrivateKey) {
    // private key stream
    privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(keyPair),
	}

    // create file 
    f, err := os.Create(fileName)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }

    err = pem.Encode(f, privateKeyBlock)
    if err != nil {
        fmt.Println("Error connecting mongo: ", err.Error())
        os.Exit(1)
    }
}

func savePublicKey(fileName string, keyPair *rsa.PrivateKey) {
    // public key stream
    asn1Bytes, err := asn1.Marshal(keyPair.PublicKey)
    publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

    // create file 
    f, err := os.Create(fileName)
    if err != nil {
        fmt.Println("Error connecting mongo: ", err.Error())
        os.Exit(1)
    }

    err = pem.Encode(f, publicKeyBlock)
    if err != nil {
        fmt.Println("Error connecting mongo: ", err.Error())
        os.Exit(1)
    }
}

func getKeySwitchKey() {

}
