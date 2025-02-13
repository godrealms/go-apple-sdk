package AppStoreNotifications

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/godrealms/go-apple-sdk/types"
	"math/big"
	"strings"
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

type SignedPayload string

// DecodedPayload Decrypt structure contents
func (sp SignedPayload) DecodedPayload() (*ResponseBodyV2DecodedPayload, error) {
	// Delimiter information
	header, payloadBytes, signature, err := sp.parseSignedPayload()
	if err != nil {
		return nil, err
	}

	// Get public key information
	certificate, err := header.X5c.GetPublicKey()
	if err != nil {
		return nil, err
	}

	var payload ResponseBodyV2DecodedPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload JSON: %v", err)
	}

	singPayload := string(sp)
	signedContent := singPayload[:strings.LastIndex(singPayload, ".")]

	// Create a hash of the signed content
	hash := sha256.Sum256([]byte(signedContent))

	// Verify the signature
	switch pubKey := certificate.PublicKey.(type) {
	case *ecdsa.PublicKey: // 使用 ECDSA 验证签名
		var r, s big.Int
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])
		if ecdsa.Verify(pubKey, hash[:], &r, &s) {
			return &payload, nil
		} else if ecdsa.VerifyASN1(pubKey, hash[:], signature) {
			return &payload, nil
		}
		return nil, fmt.Errorf("signatureBytes verification failed")
	case *rsa.PublicKey: // 使用 RSA 验证签名
		if err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature); err != nil {
			return nil, fmt.Errorf("signature verification failed: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pubKey)
	}

	return &payload, nil
}

// Parse SignedPayload and return Header, Payload and Signature
func (sp SignedPayload) parseSignedPayload() (*types.JWSDecodedHeader, []byte, []byte, error) {
	// Split signedPayload
	parts := strings.Split(string(sp), ".")
	if len(parts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid signedPayload format: expected 3 parts, got %d", len(parts))
	}

	// Decode Header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header types.JWSDecodedHeader
	if err = json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	// 解码 Payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	// 解码 Signature
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	return &header, payloadBytes, signatureBytes, nil
}

// NotificationsResponseBodyV2 The response body the App Store sends in a version 2 server notification.
type NotificationsResponseBodyV2 struct {
	// A cryptographically signed payload, in JSON Web Signature (JWS) format, that contains the response body for a version 2 notification.
	SignedPayload SignedPayload `json:"signedPayload"`
}

func Notifications(signedPayload string) (*ResponseBodyV2DecodedPayload, error) {
	notificationsResponseBodyV2 := NotificationsResponseBodyV2{
		SignedPayload: SignedPayload(signedPayload),
	}
	return notificationsResponseBodyV2.SignedPayload.DecodedPayload()
}
