package AppStoreNotifications

import (
	"github.com/godrealms/go-apple-sdk/jws"
	"github.com/godrealms/go-apple-sdk/types"
)

// App Store Server Notifications
// Monitor In-App Purchase events in real time and learn of unreported external purchase tokens,
// with server notifications from the App Store.

type Data struct {
	// The unique identifier of the app that the notification applies to.
	// This property is available for apps that users download from the App Store.
	// It isn’t present in the sandbox environment.
	AppAppleId types.AppAppleId `json:"appAppleId"`

	// The bundle identifier of the app.
	BundleId types.BundleId `json:"bundleId"`

	// The version of the build that identifies an iteration of the bundle.
	BundleVersion types.BundleVersion `json:"bundleVersion"`

	// The reason the customer requested the refund.
	// This field appears only for CONSUMPTION_REQUEST notifications,
	// which the server sends when a customer initiates a refund request
	// for a consumable in-app purchase or auto-renewable subscription.
	ConsumptionRequestReason types.ConsumptionRequestReason `json:"consumptionRequestReason"`

	// The server environment that the notification applies to, either sandbox or production.
	Environment types.Environment `json:"environment"`

	// Subscription renewal information signed by the App Store,
	// in JSON Web Signature (JWS) format. This field appears only for notifications
	// that apply to auto-renewable subscriptions.
	SignedRenewalInfo types.JWSRenewalInfo `json:"signedRenewalInfo"`

	// Transaction information signed by the App Store, in JSON Web Signature (JWS) format.
	SignedTransactionInfo types.JWSTransaction `json:"signedTransactionInfo"`

	// The status of an auto-renewable subscription as of the signedDate in the responseBodyV2DecodedPayload.
	// This field appears only for notifications sent for auto-renewable subscriptions.
	Status types.Status `json:"status"`
}

// The payload data that contains an external purchase token.
type externalPurchaseToken struct {
	// The unique identifier of the token. Use this value to report tokens and their associated transactions in the Send External Purchase Report endpoint.
	ExternalPurchaseId types.ExternalPurchaseId `json:"externalPurchaseId"`

	// The UNIX time, in milliseconds, when the system created the token.
	TokenCreationDate types.Timestamp `json:"tokenCreationDate"`

	// The app Apple ID for which the system generated the token.
	AppAppleId types.AppAppleId `json:"appAppleId"`

	// The bundle ID of the app for which the system generated the token.
	BundleId types.BundleId `json:"bundleId"`
}

type ResponseBodyV2DecodedPayload struct {
	// The in-app purchase event for which the App Store sends this version 2 notification.
	NotificationType types.NotificationType `json:"notificationType"`
	// Additional information that identifies the notification event.
	// The subtype field is present only for specific version 2 notifications.
	Subtype types.Subtype `json:"subtype"`
	// The object that contains the app metadata and signed renewal and transaction information.
	// The data, summary, and externalPurchaseToken fields are mutually exclusive.
	// The payload contains only one of these fields.
	Data Data `json:"data"`
	// The summary data that appears when the App Store server completes your request to extend a subscription renewal date for eligible subscribers.
	// For more information, see Extend Subscription Renewal Dates for All Active Subscribers.
	// The data, summary, and externalPurchaseToken fields are mutually exclusive.
	// The payload contains only one of these fields.
	Summary types.Summary `json:"summary"`
	// This field appears when the notificationType is ExternalPurchaseToken.
	// The data, summary, and externalPurchaseToken fields are mutually exclusive.
	// The payload contains only one of these fields.
	ExternalPurchaseToken externalPurchaseToken `json:"externalPurchaseToken"`
	// The App Store AppStoreServerAPI Notification version number, "2.0".
	Version types.Version `json:"version"`
	// The UNIX time, in milliseconds, that the App Store signed the JSON Web Signature data.
	SignedDate types.Timestamp `json:"signedDate"`
	// A unique identifier for the notification. Use this value to identify a duplicate notification.
	NotificationUUID types.UUID `json:"notificationUUID"`
}

// SignedPayload is the JWS-encoded V2 notification body Apple
// posts to your webhook. DecodedPayload verifies the JWS chain
// and signature using the package-default Verifier and returns
// the decoded payload; use DecodedPayloadWith for a custom
// *jws.Verifier (e.g. integration tests with self-signed certs).
type SignedPayload string

// DecodedPayload verifies the JWS chain + signature and returns
// the decoded notification payload. Returns *jws.VerificationError
// on failure.
func (sp SignedPayload) DecodedPayload() (*ResponseBodyV2DecodedPayload, error) {
	return jws.VerifyAndDecode[ResponseBodyV2DecodedPayload](
		jws.DefaultVerifier(), string(sp))
}

// DecodedPayloadWith verifies using the supplied Verifier.
func (sp SignedPayload) DecodedPayloadWith(v *jws.Verifier) (*ResponseBodyV2DecodedPayload, error) {
	return jws.VerifyAndDecode[ResponseBodyV2DecodedPayload](v, string(sp))
}

// NotificationsResponseBodyV2 The response body the App Store sends in a version 2 server notification.
type NotificationsResponseBodyV2 struct {
	// A cryptographically signed payload, in JSON Web Signature (JWS) format, that contains the response body for a version 2 notification.
	SignedPayload SignedPayload `json:"signedPayload"`
}

// Notifications is the convenience entry point: takes a raw
// signedPayload string and returns the decoded payload after
// full chain validation via DefaultVerifier.
func Notifications(signedPayload string) (*ResponseBodyV2DecodedPayload, error) {
	return SignedPayload(signedPayload).DecodedPayload()
}
