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
	subscriptions, err := AppStoreServer.GetAllSubscriptionStatuses(client, transactionId)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Environment: ", subscriptions.Environment)
	log.Println("AppAppleId: ", subscriptions.AppAppleId)
	log.Println("BundleId: ", subscriptions.BundleId)
	for _, datum := range subscriptions.Data {
		log.Println("SubscriptionGroupIdentifier: ", datum.SubscriptionGroupIdentifier)
		for _, transaction := range datum.LastTransactions {
			log.Println("OriginalTransactionId: ", transaction.OriginalTransactionId)
			log.Println("Status: ", transaction.Status)
			payload, payloadErr := transaction.SignedRenewalInfo.Decrypt()
			if payloadErr != nil {
				log.Fatal(payloadErr)
				return
			}
			log.Printf("SignedRenewalInfo: %+v\n", payload)
			decrypt, decryptErr := transaction.SignedTransactionInfo.Decrypt()
			if decryptErr != nil {
				log.Fatal(decryptErr)
				return
			}
			log.Printf("SignedTransactionInfo: %+v\n", decrypt)
		}
	}
}
