package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

// RefundHistoryResponse A response that contains an array of signed JSON Web Signature (JWS) refunded transactions,
// and paging information.
type RefundHistoryResponse struct {
	// A Boolean value that indicates whether the App Store has more transactions than it returns in signedTransactions.
	// If the value is true, use the revision token to request the next set of transactions by calling Get Refund History.
	HasMore types.HasMore `json:"hasMore"`

	// A token you provide in a query to request the next set of transactions from the Get Refund History endpoint.
	Revision types.Revision `json:"revision"`

	// A list of up to 20 JWS transactions, or an empty array if the customer hasn’t received any refunds in your app.
	// The transactions are sorted in ascending order by revocationDate.
	SignedTransactions []types.JWSTransaction `json:"signedTransactions"`
}

// GetRefundHistory Get a paginated list of all of a customer’s refunded in-app purchases for your app.
func GetRefundHistory(client *Apple.Client, transactionId string) (*RefundHistoryResponse, error) {
	client.SetService(Apple.AppStoreServerClient)
	var result = new(RefundHistoryResponse)
	params := Apple.RequestParams{
		Method: "GET",
		Path:   "/inApps/v2/refund/lookup/{transactionId}",
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
