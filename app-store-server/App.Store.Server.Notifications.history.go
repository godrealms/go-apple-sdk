package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
)

type NotificationHistoryResponse struct{}

// GetNotificationHistory Get a list of notifications that the App Store server attempted to send to your server.
// paginationToken: A pagination token that you return to the endpoint on a subsequent call to receive the next set of results.
func GetNotificationHistory(client *Apple.Client, paginationToken string) (*NotificationHistoryResponse, error) {
	var result = new(NotificationHistoryResponse)
	client.SetService(Apple.AppStoreServerClient)
	params := Apple.RequestParams{
		Method: "POST",
		Path:   "/inApps/v1/notifications/history",
		Result: result,
		Body: map[string]string{
			"paginationToken": paginationToken,
		},
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}
	if err := client.Request(params); err != nil {
		return nil, err
	}
	return result, nil
}
