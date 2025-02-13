package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

type OrderLookupResponse struct {
	// The status that indicates whether the order ID is valid.
	Status types.OrderLookupStatus `json:"status"`
	// An array of in-app purchase transactions that are part of order, signed by Apple, in JSON Web Signature format.
	SignedTransactions []types.JWSTransaction `json:"signedTransactions"`
}

// LookUpOrderID Get a customer’s in-app purchases from a receipt using the order ID.
func LookUpOrderID(client *Apple.Client, orderId string) (*OrderLookupResponse, error) {
	client.SetService(Apple.AppStoreServerClient)
	var result = new(OrderLookupResponse)
	params := Apple.RequestParams{
		Method: "GET",
		Path:   "/inApps/v1/lookup/{orderId}",
		Result: result,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		PathParams: map[string]string{
			"orderId": orderId,
		},
	}
	if err := client.Request(params); err != nil {
		return nil, err
	}
	return result, nil
}
