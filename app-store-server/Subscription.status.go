package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

// LastTransactionsItem
// The most recent App Store-signed transaction information and App Store-signed renewal information for an auto-renewable subscription.
type LastTransactionsItem struct {
	// The original transaction identifier of the auto-renewable subscription.
	OriginalTransactionId types.OriginalTransactionId `json:"originalTransactionId"`
	// The status of the auto-renewable subscription.
	Status types.Status `json:"status"`
	// The subscription renewal information signed by the App Store, in JSON Web Signature (JWS) format.
	SignedRenewalInfo types.JWSRenewalInfo `json:"signedRenewalInfo"`
	// The transaction information signed by the App Store, in JWS format.
	SignedTransactionInfo types.JWSTransaction `json:"signedTransactionInfo"`
}

// SubscriptionGroupIdentifierItem
// Information for auto-renewable subscriptions,
// including signed transaction information and signed renewal information,
// for one subscription group.
type SubscriptionGroupIdentifierItem struct {
	// The subscription group identifier of the auto-renewable subscriptions in the lastTransactions array.
	SubscriptionGroupIdentifier types.SubscriptionGroupIdentifier `json:"subscriptionGroupIdentifier"`
	// An array of the most recent App Store-signed transaction information and App Store-signed renewal
	// information for all auto-renewable subscriptions in the subscription group.
	LastTransactions []LastTransactionsItem `json:"lastTransactions"`
}

// StatusResponse
// A response that contains status information for all of a customer’s auto-renewable subscriptions in your app.
type StatusResponse struct {
	// An array of information for auto-renewable subscriptions, including App Store-signed transaction information and App Store-signed renewal information.
	Data []SubscriptionGroupIdentifierItem `json:"data"`
	// The server environment, sandbox or production, in which the App Store generated the response.
	Environment types.Environment `json:"environment"`
	// Your app’s App Store identifier.
	AppAppleId types.AppAppleId `json:"appAppleId"`
	// Your app’s bundle identifier.
	BundleId types.BundleId `json:"bundleId"`
}

// GetAllSubscriptionStatuses
// Get the statuses for all of a customer’s auto-renewable subscriptions in your app.
func GetAllSubscriptionStatuses(client *Apple.Client, transactionId string) (*StatusResponse, error) {
	client.SetService(Apple.AppStoreServerClient)
	var result = new(StatusResponse)
	params := Apple.RequestParams{
		Method: "GET",
		Path:   "/inApps/v1/subscriptions/{transactionId}",
		Result: result,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		PathParams: map[string]string{
			"transactionId": transactionId,
		},
	}
	if err := client.Request(params); err != nil {
		return nil, err
	}
	return result, nil
}
