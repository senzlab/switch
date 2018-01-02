package main

import (
    "crypto/rand"
    "crypto"
    "crypto/rsa"
    "crypto/x509"
    "crypto/sha256"
    "encoding/asn1"
    "encoding/pem"
    "encoding/base64"
    "io/ioutil"
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
    keyPair := initKeyPair()
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
        fmt.Println("Error mongo: ", err.Error())
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
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }

    err = pem.Encode(f, publicKeyBlock)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }
}

func getSwitchKey()*rsa.PrivateKey {
    keyData, err := ioutil.ReadFile(config.idRsa)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }

    keyBlock, _ := pem.Decode(keyData)
    if keyBlock == nil {
        fmt.Println("Error : ", "invalid key")
        os.Exit(1)
    }

    privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
    if err != nil {
        fmt.Println("Error : ", "invalid key")
        os.Exit(1)
    }

    return privateKey
}

func getSenzieKey(keyStr string) *rsa.PublicKey {
    key := "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCTe3NEX8omY+NOcrvtZ27cUR6HJL5JyVj8lv2RmIa4XicoxD9NuPkMeovDLOyhAqUITosbZCmdGjKfyrLzGxU33VlyPjODfFars3baLf4Hdh7IfN6Z+w5xevE88hPaJdWXnZyfvtUHxGF/0mOsowfOZT5Cm+6+G6PPGDdAnQ2P6wIDAQAB"
	data, err := base64.StdEncoding.DecodeString(key)
    if err != nil {
        println(err.Error())
        return nil
    }

    pub, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		println(err.Error())
	}
    switch pub := pub.(type) {
    case *rsa.PublicKey:
        return pub
    default:
        return nil
    }
}

func sing(payload string, key *rsa.PrivateKey) (string, error) {
    // hash the payload first
	hashed := sha256.Sum256([]byte(payload))

    // sign the hased payload
    signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
        return "", err
    }

    // reutrn base64 encoded string
    return base64.StdEncoding.EncodeToString(signature), nil
}

func verify(payload string, signedPayload string, key *rsa.PublicKey) error {
    // decode base64 signed payload
    signature, err := base64.StdEncoding.DecodeString(signedPayload)
    if err != nil {
        return err
    }

    // hased payload
	hashed := sha256.Sum256([]byte(payload))

    return rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], signature) 
}

func getPrivateKey(keyStr string) (*rsa.PrivateKey) {
    key := "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCTe3NEX8omY+NOcrvtZ27cUR6HJL5JyVj8lv2RmIa4XicoxD9NuPkMeovDLOyhAqUITosbZCmdGjKfyrLzGxU33VlyPjODfFars3baLf4Hdh7IfN6Z+w5xevE88hPaJdWXnZyfvtUHxGF/0mOsowfOZT5Cm+6+G6PPGDdAnQ2P6wIDAQAB"
	data, err := base64.StdEncoding.DecodeString(key)
    if err != nil {
        println(err.Error())
        return nil
    }

    privateKey, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		println(err.Error())
	}

    return privateKey
}
