package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

type SendTestNotificationResponse struct {
	TestNotificationToken string `json:"testNotificationToken"`
}

type SendAttemptItem struct {
	// The date the App Store server attempts to send the notification.
	AttemptDate types.Timestamp `json:"attemptDate"`
	// The success or error information the App Store server records when it attempts to send an App Store server notification to your server.
	SendAttemptResult string `json:"sendAttemptResult"`
}

type CheckTestNotificationResponse struct {
	// An array of information the App Store server records for its attempts to send the TEST notification to your server.
	// The array may contain a maximum of six sendAttemptItem objects.
	SendAttempts []SendAttemptItem `json:"sendAttempts"`
	// The signed payload, in JWS format,
	// that contains the TEST notification that the App Store server sent to your server.
	SignedPayload string `json:"signedPayload"`
}

// RequestTestNotification Ask App Store AppStoreServerAPI Notifications to send a test notification to your server.
func RequestTestNotification(client *Apple.Client) (*SendTestNotificationResponse, error) {
	var result = new(SendTestNotificationResponse)
	client.SetService(Apple.AppStoreServerClient)
	params := Apple.RequestParams{
		Method: "POST",
		Path:   "/inApps/v1/notifications/test",
		Result: result,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}
	if err := client.Request(params); err != nil {
		return nil, err
	}
	return result, nil

}

// GetTestNotificationStatus Check the status of the test App Store server notification sent to your server.
func GetTestNotificationStatus(client *Apple.Client, testNotificationToken string) (*CheckTestNotificationResponse, error) {
	var result = new(CheckTestNotificationResponse)
	client.SetService(Apple.AppStoreServerClient)
	params := Apple.RequestParams{
		Method: "GET",
		Path:   "/inApps/v1/notifications/test/{testNotificationToken}",
		Result: result,
		Headers: map[string]string{
			"Accept": "application/json",
		},
		PathParams: map[string]string{
			"testNotificationToken": testNotificationToken,
		},
	}

	if err := client.Request(params); err != nil {
		return nil, err
	}
	return result, nil
}
