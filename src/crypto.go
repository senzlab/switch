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
    "fmt"
    "os"
)

var keySize = 1024

func setUpKeys() {
    if _, e1 := os.Stat(config.dotKeys); e1 != nil {
        // keys directory not exists
        // create it
        e2 := os.Mkdir(config.dotKeys, 0700)
        if e2 != nil {
            fmt.Println("Error : ", e2.Error())
            os.Exit(1)
        }
    }

    if _, e3 := os.Stat(config.idRsa); e3 != nil {
        // keys not exists
        // generate key pair
        // save private key
        // save public key
        keyPair := initKeyPair()
        saveIdRsa(config.idRsa, keyPair)
        saveIdRsaPub(config.idRsaPub, keyPair)
    }
}

func initKeyPair() *rsa.PrivateKey {
    // generate key pair
    keyPair, e1 := rsa.GenerateKey(rand.Reader, keySize)
    if e1 != nil {
        fmt.Println("Error : ", e1.Error())
        os.Exit(1)
    }

    // validate key
    e2 := keyPair.Validate()
	if e2 != nil {
        fmt.Println("Error : ", e2.Error())
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
    f, e1 := os.Create(fileName)
    if e1 != nil {
        fmt.Println("Error : ", e1.Error())
        os.Exit(1)
    }

    e2 := pem.Encode(f, privateKeyBlock)
    if e2 != nil {
        fmt.Println("Error mongo: ", e2.Error())
        os.Exit(1)
    }
}

func saveIdRsaPub(fileName string, keyPair *rsa.PrivateKey) {
    // public key stream
    pubKeyBytes, e1 := x509.MarshalPKIXPublicKey(&keyPair.PublicKey)
    if e1 != nil {
        fmt.Println(e1.Error())
        os.Exit(1)
    }

    publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

    // create file
    f, e2 := os.Create(fileName)
    if e2 != nil {
        fmt.Println("Error : ", e2.Error())
        os.Exit(1)
    }

    e3 := pem.Encode(f, publicKeyBlock)
    if e3 != nil {
        fmt.Println(e3.Error())
        os.Exit(1)
    }
}

func getIdRsa()*rsa.PrivateKey {
    keyData, e1 := ioutil.ReadFile(config.idRsa)
    if e1 != nil {
        fmt.Println(e1.Error())
        os.Exit(1)
    }

    keyBlock, _ := pem.Decode(keyData)
    if keyBlock == nil {
        fmt.Println("Error : ", "invalid key")
        os.Exit(1)
    }

    privateKey, e2 := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
    if e2 != nil {
        fmt.Println(e2.Error())
        os.Exit(1)
    }

    return privateKey
}

func getIdRsaPub()*rsa.PublicKey {
    keyData, e1 := ioutil.ReadFile(config.idRsaPub)
    if e1 != nil {
        fmt.Println(e1.Error())
        return nil
    }

    keyBlock, _ := pem.Decode(keyData)
    if keyBlock == nil {
        fmt.Println("Error : ", "invalid key")
        return nil
    }

    publicKey, e2 := x509.ParsePKIXPublicKey(keyBlock.Bytes)
    if e2 != nil {
        fmt.Println(e2.Error())
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
    keyData, e1 := ioutil.ReadFile(config.idRsaPub)
    if e1 != nil {
        fmt.Println(e1.Error())
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

func getSenzieRsa(keyStr string) *rsa.PrivateKey {
    // key is base64 encoded
	data, e1 := base64.StdEncoding.DecodeString(keyStr)
    if e1 != nil {
        println(e1.Error())
        return nil
    }

    // get rsa private key
    key, e2 := x509.ParsePKCS8PrivateKey(data)
	if e2 != nil {
        println(e2.Error())
        return nil
	}
    switch key := key.(type) {
    case *rsa.PrivateKey:
        return key
    default:
        return nil
    }

    return nil
}

func getSenzieRsaPub(keyStr string) *rsa.PublicKey {
    // key is base64 encoded
	data, e1 := base64.StdEncoding.DecodeString(keyStr)
    if e1 != nil {
        println(e1.Error())
        return nil
    }

    // get rsa public key
    pub, e2 := x509.ParsePKIXPublicKey(data)
	if e2 != nil {
		println(e2.Error())
        return nil
	}
    switch pub := pub.(type) {
    case *rsa.PublicKey:
        return pub
    default:
        return nil
    }
}

func sign(payload string, key *rsa.PrivateKey) (string, error) {
    // remove unwated characters and get sha256 hash of the payload
	hashed := sha256.Sum256([]byte(formatToSign(payload)))

    // sign the hased payload
    signature, e1 := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
    if e1 != nil {
		println(e1.Error())
        return "", e1
    }

    // reutrn base64 encoded string
    return base64.StdEncoding.EncodeToString(signature), nil
}

func verify(payload string, signature64 string, key *rsa.PublicKey) error {
    // decode base64 encoded signature
    signature, e1 := base64.StdEncoding.DecodeString(signature64)
    if e1 != nil {
        println(e1.Error())
        return e1
    }

    // remove unwated characters and get sha256 hash of the payload
	hashed := sha256.Sum256([]byte(formatToSign(payload)))

    return rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], signature)
}
