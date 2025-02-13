package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

type TransactionInfoResponse struct {
	// A customer’s in-app purchase transaction, signed by Apple, in JSON Web Signature (JWS) format.
	SignedTransactionInfo types.JWSTransaction `json:"signedTransactionInfo"`
}

func GetTransactionInfo(client *Apple.Client, transactionId string) (*TransactionInfoResponse, error) {
	client.SetService(Apple.AppStoreServerClient)
	var result = new(TransactionInfoResponse)
	params := Apple.RequestParams{
		Method: "GET",
		Path:   "/inApps/v1/transactions/{transactionId}",
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
