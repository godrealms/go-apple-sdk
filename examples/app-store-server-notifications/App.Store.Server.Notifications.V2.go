package main

import (
	AppStoreNotifications "github.com/godrealms/go-apple-sdk/app-store-server-notifications"
	"log"
)

func main() {
	signedPayload := ""
	notifications, err := AppStoreNotifications.Notifications(signedPayload)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("NotificationType: ", notifications.NotificationType)
	log.Println("Subtype: ", notifications.Subtype)
	log.Println("Summary: ", notifications.Summary)
	log.Println("Version: ", notifications.Version)
	log.Println("SignedDate: ", notifications.SignedDate)
	log.Println("NotificationUUID: ", notifications.NotificationUUID)

	log.Println("ExternalPurchaseToken.ExternalPurchaseId: ", notifications.ExternalPurchaseToken.ExternalPurchaseId)
	log.Println("ExternalPurchaseToken.TokenCreationDate: ", notifications.ExternalPurchaseToken.TokenCreationDate)
	log.Println("ExternalPurchaseToken.AppAppleId: ", notifications.ExternalPurchaseToken.AppAppleId)
	log.Println("ExternalPurchaseToken.BundleId: ", notifications.ExternalPurchaseToken.BundleId)

	log.Println("Data.AppAppleId: ", notifications.Data.AppAppleId)
	log.Println("Data.BundleId: ", notifications.Data.BundleId)
	log.Println("Data.BundleVersion: ", notifications.Data.BundleVersion)
	log.Println("Data.ConsumptionRequestReason: ", notifications.Data.ConsumptionRequestReason)
	log.Println("Data.Environment: ", notifications.Data.Environment)
	log.Println("Data.Status: ", notifications.Data.Status)

	renewalInfo, renewalInfoErr := notifications.Data.SignedRenewalInfo.Decrypt()
	if err != nil {
		log.Fatal(renewalInfoErr)
	}
	log.Printf("Data.SignedRenewalInfo: %+v\n", renewalInfo)

	transaction, transactionErr := notifications.Data.SignedTransactionInfo.Decrypt()
	if err != nil {
		log.Fatal(transactionErr)
	}
	log.Printf("Data.SignedTransactionInfo: %+v\n", transaction)

}
