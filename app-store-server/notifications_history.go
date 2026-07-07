package AppStoreServer

import (
	"context"

	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

// NotificationHistoryRequest holds the filter criteria for
// GetNotificationHistory. StartDate and EndDate are required by Apple; the
// remaining fields are optional filters.
type NotificationHistoryRequest struct {
	// The start date of the timespan for the requested notification history
	// records, as a UNIX time in milliseconds. Required.
	StartDate types.Timestamp `json:"startDate"`
	// The end date of the timespan for the requested notification history
	// records, as a UNIX time in milliseconds. Required.
	EndDate types.Timestamp `json:"endDate"`
	// Optional: filter to a single notification type.
	NotificationType string `json:"notificationType,omitempty"`
	// Optional: filter to a single notification subtype.
	NotificationSubtype string `json:"notificationSubtype,omitempty"`
	// Optional: when true, return only notifications that failed to reach
	// your server.
	OnlyFailures bool `json:"onlyFailures,omitempty"`
	// Optional: filter to notifications for a single transaction.
	TransactionId string `json:"transactionId,omitempty"`
}

// NotificationHistoryResponseItem is a single entry in the notification
// history: the signed notification plus the App Store server's send attempts.
type NotificationHistoryResponseItem struct {
	// The signed payload, in JWS format, of the notification.
	SignedPayload string `json:"signedPayload"`
	// The App Store server's attempts to send the notification to your server.
	SendAttempts []SendAttemptItem `json:"sendAttempts"`
}

// NotificationHistoryResponse is the App Store server's paginated notification
// history response.
type NotificationHistoryResponse struct {
	// The notification history records that match the request criteria.
	NotificationHistory []NotificationHistoryResponseItem `json:"notificationHistory"`
	// A Boolean value that indicates whether more records are available.
	HasMore types.HasMore `json:"hasMore"`
	// A pagination token to pass on a subsequent call to receive the next set
	// of results.
	PaginationToken string `json:"paginationToken"`
}

// GetNotificationHistory gets a list of notifications the App Store server
// attempted to send to your server.
//
// The filter criteria (startDate/endDate/etc.) travel in the request body,
// while paginationToken — returned from a previous call — is passed as a query
// parameter to fetch the next page. Pass an empty paginationToken for the
// first page.
func GetNotificationHistory(ctx context.Context, client *Apple.Client, request NotificationHistoryRequest, paginationToken string) (*NotificationHistoryResponse, error) {
	var result = new(NotificationHistoryResponse)
	client.SetService(Apple.AppStoreServerClient)

	query := map[string]any{}
	if paginationToken != "" {
		query["paginationToken"] = paginationToken
	}

	params := Apple.RequestParams{
		Ctx:         ctx,
		Method:      "POST",
		Path:        "/inApps/v1/notifications/history",
		Result:      result,
		Body:        request,
		QueryParams: query,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
	}
	if err := client.Request(params); err != nil {
		return nil, err
	}
	return result, nil
}
