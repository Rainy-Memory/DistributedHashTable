package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"unsafe"
)

const (
	MaxEncryptedSize = 115
	RSABits          = 1024
	PackNumber       = 7 // 887 / 115 == 7, 887 is length of 1024 bits RSA key
	MaxMessageLength = 90
)

func GetLocalAddress() string {
	var localAddress string
	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}
	// find the first non-loopback interface with an IP address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localAddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localAddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}
	return localAddress
}

func GenerateRSAKey(bits int) (prvKey, pubKey []byte) {
	privateKey, err0 := rsa.GenerateKey(rand.Reader, bits)
	if err0 != nil {
		panic(err0)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvKey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err1 := x509.MarshalPKIXPublicKey(publicKey)
	if err1 != nil {
		panic(err1)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey = pem.EncodeToMemory(block)
	return
}

func Encrypt(data, pubKey []byte) []byte {
	block, _ := pem.Decode(pubKey)
	if block == nil {
		panic(errors.New("public key error"))
	}
	pubInterface, err0 := x509.ParsePKIXPublicKey(block.Bytes)
	if err0 != nil {
		panic(err0)
	}
	pub := pubInterface.(*rsa.PublicKey)
	ciphertext, err1 := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err1 != nil {
		panic(err1)
	}
	return ciphertext
}

func Decrypt(ciphertext, prvKey []byte) []byte {
	block, _ := pem.Decode(prvKey)
	if block == nil {
		panic(errors.New("private key error"))
	}
	prv, err0 := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err0 != nil {
		panic(err0)
	}
	data, err1 := rsa.DecryptPKCS1v15(rand.Reader, prv, ciphertext)
	if err1 != nil {
		panic(err1)
	}
	return data
}

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
