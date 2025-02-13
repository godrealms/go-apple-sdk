package main

import (
	Apple "github.com/godrealms/go-apple-sdk"
	AppStoreServer "github.com/godrealms/go-apple-sdk/app-store-server"
	"log"
)

func main() {
	kid := ""           // Your private key ID from App Store Connect (Ex: 2X9R4HXF34)
	iss := ""           // Your issuer ID from the Keys page in App Store Connect (Ex: “57246542-96fe-1a63-e053-0824d011072a")
	bid := ""           // Your app’s bundle ID (Ex: “com.example.testbundleid”)
	privateKey := ""    // Your private key
	transactionId := "" // Transaction ID
	client := Apple.NewClient(true, kid, iss, bid, privateKey)
	info, err := AppStoreServer.GetTransactionInfo(client, transactionId)
	if err != nil {
		log.Fatal(err)
		return
	}
	decrypt, decryptErr := info.SignedTransactionInfo.Decrypt()
	if decryptErr != nil {
		log.Fatal(decryptErr)
		return
	}
	log.Printf("%+v\n", decrypt)
}
