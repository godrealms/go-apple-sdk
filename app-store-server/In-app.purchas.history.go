package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

type HistoryResponse struct {
	// The app’s identifier in the App Store.
	AppAppleId types.AppAppleId `json:"appAppleId"`
	// The bundle identifier of the app.
	BundleId types.BundleId `json:"bundleId"`
	// The server environment in which you’re making the request, whether sandbox or production.
	Environment types.Environment `json:"environment"`
	// A Boolean value that indicates whether the App Store has more transactions than it returns in this response.
	// If the value is true, use the revision token to request the next set of transactions.
	HasMore types.HasMore `json:"hasMore"`
	// A token you use in a query to request the next set of transactions from the endpoint.
	Revision types.Revision `json:"revision"`
	// An array of in-app purchase transactions for the customer, signed by Apple, in JSON Web Signature (JWS) format.
	SignedTransactions []types.JWSTransaction `json:"signedTransactions"`
}

// GetTransactionHistory Get a customer’s in-app purchase transaction history for your app.
func GetTransactionHistory(client *Apple.Client, transactionId string, queryParams ...map[string]any) (*HistoryResponse, error) {
	var result = new(HistoryResponse)
	client.SetService(Apple.AppStoreServerClient)
	params := Apple.RequestParams{
		Method: "GET",
		Path:   "/inApps/v2/history/{transactionId}",
		Result: result,
		Headers: map[string]string{
			"Accept": "application/json",
		},
		PathParams: map[string]string{
			"transactionId": transactionId,
		},
	}

	if len(queryParams) > 0 && queryParams[0] != nil {
		params.QueryParams = queryParams[0]
	}

	if err := client.Request(params); err != nil {
		return nil, err
	}
	return result, nil
}
