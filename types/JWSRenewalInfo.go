package types

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// The identifier of the product that renews at the next billing period.
type autoRenewProductId string

// The renewal status for an auto-renewable subscription.
// 0: Automatic renewal is off. The customer has turned off automatic renewal for the subscription, and it won’t renew at the end of the current subscription period.
// 1: Automatic renewal is on. The subscription renews at the end of the current subscription period.
type autoRenewStatus int32

// An array of win-back offer identifiers that a customer is eligible to redeem, which sorts the identifiers with the best offers first.
type eligibleWinBackOfferIds []offerIdentifier

// The reason an auto-renewable subscription expired.
// 1: The customer canceled their subscription.
// 2: Billing error; for example, the customer’s payment information is no longer valid.
// 3: The customer didn’t consent to an auto-renewable subscription price increase that requires customer consent, allowing the subscription to expire.
// 4: The product wasn’t available for purchase at the time of renewal.
// 5: The subscription expired for some other reason.
type expirationIntent int32

// A Boolean value that indicates whether the App Store is attempting to automatically renew an expired subscription.
type isInBillingRetryPeriod bool

// The status that indicates whether an auto-renewable subscription is subject to a price increase.
// 0: The customer hasn’t yet responded to an auto-renewable subscription price increase that requires customer consent.
// 1: The customer consented to an auto-renewable subscription price increase that requires customer consent,
// or the App Store has notified the customer of an auto-renewable subscription price increase that doesn’t require consent.
type priceIncreaseStatus int32

// The renewal price, in milliunits, of the auto-renewable subscription that renews at the next billing period.
type renewalPrice int64

// JWSRenewalInfoDecodedPayload A decoded payload containing subscription renewal information for an auto-renewable subscription.
type JWSRenewalInfoDecodedPayload struct {
	// The identifier of the product that renews at the next billing period.
	AutoRenewProductId autoRenewProductId `json:"autoRenewProductId"`

	// The renewal status of the auto-renewable subscription.
	AutoRenewStatus autoRenewStatus `json:"autoRenewStatus"`

	// The currency code for the renewalPrice of the subscription.
	Currency currency `json:"currency"`

	// The list of win-back offer IDs that the customer is eligible for.
	EligibleWinBackOfferIds eligibleWinBackOfferIds `json:"eligibleWinBackOfferIds"`

	// The server environment, either sandbox or production.
	Environment Environment `json:"environment"`

	// The reason the subscription expired.
	ExpirationIntent expirationIntent `json:"expirationIntent"`

	// The time when the Billing Grace Period for subscription renewals expires.
	GracePeriodExpiresDate Timestamp `json:"gracePeriodExpiresDate"`

	// A Boolean value that indicates whether the App Store is attempting to automatically renew the expired subscription.
	IsInBillingRetryPeriod isInBillingRetryPeriod `json:"isInBillingRetryPeriod"`

	// The payment mode of the discount offer.
	OfferDiscountType offerDiscountType `json:"offerDiscountType"`

	// The offer code or the promotional offer identifier.
	OfferIdentifier offerIdentifier `json:"offerIdentifier"`

	// The type of subscription offer.
	OfferType offerType `json:"offerType"`

	// The transaction identifier of the original purchase associated with this transaction.
	OriginalTransactionId OriginalTransactionId `json:"originalTransactionId"`

	// The status that indicates whether the auto-renewable subscription is subject to a price increase.
	PriceIncreaseStatus priceIncreaseStatus `json:"priceIncreaseStatus"`

	// The product identifier of the In-App Purchase.
	ProductId ProductId `json:"productId"`

	// The earliest start date of the auto-renewable subscription in a series of subscription purchases that ignores all lapses of paid service that are 60 days or fewer.
	RecentSubscriptionStartDate Timestamp `json:"recentSubscriptionStartDate"`

	// The UNIX time, in milliseconds, when the most recent auto-renewable subscription purchase expires.
	RenewalDate Timestamp `json:"renewalDate"`

	// The renewal price, in milliunits, of the auto-renewable subscription that renews at the next billing period.
	RenewalPrice renewalPrice `json:"renewalPrice"`

	// The UNIX time, in milliseconds, that the App Store signed the JSON Web Signature (JWS) data.
	SignedDate Timestamp `json:"signedDate"`
}

// JWSRenewalInfo Subscription renewal information, signed by the App Store, in JSON Web Signature (JWS) format.
type JWSRenewalInfo string

func (j JWSRenewalInfo) Decrypt() (*JWSRenewalInfoDecodedPayload, error) {
	// Delimiter information
	header, payloadBytes, signature, err := j.parseSignedPayload()
	if err != nil {
		return nil, err
	}
	// Get public key information
	certificate, err := header.X5c.GetPublicKey()
	if err != nil {
		return nil, err
	}

	var payload JWSRenewalInfoDecodedPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload JSON: %v", err)
	}

	singPayload := string(j)
	signedContent := singPayload[:strings.LastIndex(singPayload, ".")]

	// Create a hash of the signed content
	hash := sha256.Sum256([]byte(signedContent))

	// Verify the signatureBytes
	switch pubKey := certificate.PublicKey.(type) {
	case *ecdsa.PublicKey:
		// 使用 ECDSA 验证签名
		var r, s big.Int
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])
		if ecdsa.Verify(pubKey, hash[:], &r, &s) {
			return &payload, nil
		} else if ecdsa.VerifyASN1(pubKey, hash[:], signature) {
			return &payload, nil
		}
		return nil, fmt.Errorf("signatureBytes verification failed")
	case *rsa.PublicKey:
		err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature)
		if err != nil {
			return nil, fmt.Errorf("signatureBytes verification failed: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pubKey)
	}
	return &payload, nil
}

// Parse signedPayload and return Header, Payload and Signature
func (j JWSRenewalInfo) parseSignedPayload() (*JWSDecodedHeader, []byte, []byte, error) {
	// Split signedPayload
	parts := strings.Split(string(j), ".")
	if len(parts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid signedPayload format: expected 3 parts, got %d", len(parts))
	}

	// Decode Header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header JWSDecodedHeader
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
