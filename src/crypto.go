package main

import (
    "crypto/rand"
    "crypto"
    "crypto/rsa"
    "crypto/x509"
    "crypto/sha256"
    "encoding/pem"
    "encoding/base64"
    "io/ioutil"
    "strings"
    "fmt"
    "os"
)

var keySize = 1024

func setUpKeys() {
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
    saveIdRsa(config.idRsa, keyPair)
    saveIdRsaPub(config.idRsaPub, keyPair)
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

func saveIdRsa(fileName string, keyPair *rsa.PrivateKey) {
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

func saveIdRsaPub(fileName string, keyPair *rsa.PrivateKey) {
    // public key stream
    pubKeyBytes, err := x509.MarshalPKIXPublicKey(&keyPair.PublicKey)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }

    publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

    // create file
    f, err := os.Create(fileName)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        os.Exit(1)
    }

    err = pem.Encode(f, publicKeyBlock)
    if err != nil {
        fmt.Println("Error mongo: ", err.Error())
        os.Exit(1)
    }
}

func getIdRsa()*rsa.PrivateKey {
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

func getIdRsaPub()*rsa.PublicKey {
    keyData, err := ioutil.ReadFile(config.idRsaPub)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        return nil
    }

    keyBlock, _ := pem.Decode(keyData)
    if keyBlock == nil {
        fmt.Println("Error : ", "invalid key")
        return nil
    }

    publicKey, err := x509.ParsePKIXPublicKey(keyBlock.Bytes)
    if err != nil {
        fmt.Println("Error : ", "invalid key")
        return nil
    }
    switch publicKey := publicKey.(type) {
    case *rsa.PublicKey:
        return publicKey
    default:
        return nil
    }
}

func getIdRsaPubStr()string {
    keyData, err := ioutil.ReadFile(config.idRsaPub)
    if err != nil {
        fmt.Println("Error : ", err.Error())
        return ""
    }

    keyBlock, _ := pem.Decode(keyData)
    if keyBlock == nil {
        fmt.Println("Error : ", "invalid key")
        return ""
    }

    // encode base64 key data
    return base64.StdEncoding.EncodeToString(keyBlock.Bytes)
}

func getSenziePubStr(keyStr string) *rsa.PublicKey {
    // key is base64 encoded
	data, err := base64.StdEncoding.DecodeString(keyStr)
    if err != nil {
        println(err.Error())
        return nil
    }

    // get rsa public key
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

func sign(payload string) (string, error) {
    // first replace unwanted characters and format payload
    // then hash it
    replacer := strings.NewReplacer(";", "", "\n", "", "\r", "")
    fPayload := strings.TrimSpace(replacer.Replace(payload))
	hashed := sha256.Sum256([]byte(fPayload))

    // sign the hased payload
    signature, err := rsa.SignPKCS1v15(rand.Reader, getIdRsa(), crypto.SHA256, hashed[:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
        return "", err
    }

    // reutrn base64 encoded string
    return base64.StdEncoding.EncodeToString(signature), nil
}

func verify(payload string, signature64 string, key *rsa.PublicKey) error {
    // decode base64 signature 
    signature, err := base64.StdEncoding.DecodeString(signature64)
    if err != nil {
        return err
    }

    // replace unwanted characters and format payload
    // then hash it
    replacer := strings.NewReplacer(";", "", "\n", "", "\r", "")
    fPayload := strings.TrimSpace(replacer.Replace(payload))
	hashed := sha256.Sum256([]byte(fPayload))

    return rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], signature)
}
