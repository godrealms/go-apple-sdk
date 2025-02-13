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
	history, err := AppStoreServer.GetTransactionHistory(client, transactionId, map[string]any{
		// A token you provide to get the next set of up to 20 transactions.
		// All responses include a revision token.
		// Use the revision token from the previous HistoryResponse.
		//"revision": "",

		// An optional start date of the timespan for the transaction history records you’re requesting.
		// The startDate needs to precede the endDate if you specify both dates.
		// The results include a transaction if its purchaseDate is equal to or greater than the startDate.
		//"startDate": "1",

		// An optional end date of the timespan for the transaction history records you’re requesting.
		// Choose an endDate that’s later than the startDate if you specify both dates.
		// Using an endDate in the future is valid.
		// The results include a transaction if its purchaseDate is less than the endDate.
		//"endDate": "20",

		// An optional filter that indicates the product identifier to include in the transaction history.
		// Your query may specify more than one productID.
		//"productId": []string{"test1"},

		// An optional filter that indicates the product type to include in the transaction history.
		// Your query may specify more than one productType.
		// Possible Values: AUTO_RENEWABLE, NON_RENEWABLE, CONSUMABLE, NON_CONSUMABLE
		//"productType": []string{"AUTO_RENEWABLE", "NON_RENEWABLE", "CONSUMABLE", "NON_CONSUMABLE"},

		// A string that describes whether the transaction was purchased by the customer,
		// or is available to them through Family Sharing.
		// Possible Values: FAMILY_SHARED, PURCHASED
		// FAMILY_SHARED: The transaction belongs to a family member who benefits from service.
		// PURCHASED: The transaction belongs to the purchaser.
		//"inAppOwnershipType": []string{"FAMILY_SHARED", "PURCHASED"},

		// An optional sort order for the transaction history records.
		// The response sorts the transaction records by their recently modified date.
		// The default value is ASCENDING, so you receive the oldest records first.
		// Possible Values: ASCENDING, DESCENDING
		"sort": "DESCENDING",

		// An optional Boolean value that indicates whether the response includes only revoked transactions when the value is true,
		// or contains only nonrevoked transactions when the value is false.
		// By default, the request doesn’t include this parameter.
		// Possible Values: true, false
		//"revoked": true,

		// An optional filter that indicates the subscription group identifier to include in the transaction history.
		// Your query may specify more than one subscriptionGroupIdentifier.
		//"subscriptionGroupIdentifier": []string{},
	})
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Printf("%+v", history)
	for _, transaction := range history.SignedTransactions {
		decrypt, decryptErr := transaction.Decrypt()
		if decryptErr != nil {
			log.Panic(decryptErr)
			return
		}
		log.Printf("%+v", decrypt)
	}
}
