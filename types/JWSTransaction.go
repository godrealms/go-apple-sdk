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

// The three-letter ISO 4217 currency code for the price of the product.
type currency string

// The Boolean value that indicates whether the customer upgraded to another subscription.
type isUpgraded bool

// The payment mode for subscription offers on an auto-renewable subscription.
// FREE_TRIAL: A payment mode of a product discount that indicates a free trial.
// PAY_AS_YOU_GO: A payment mode of a product discount that customers pay over a single or multiple billing periods.
// PAY_UP_FRONT: A payment mode of a product discount that customers pay up front.
type offerDiscountType string

// The string identifier of a subscription offer that you create in App Store AppStoreConnectAPI.
type offerIdentifier string

// The type of subscription offer.
// 1: An introductory offer.
// 2: A promotional offer.
// 3: An offer with a subscription offer code.
// 4: A win-back offer.
type offerType int32

// The price, in milliunits, of the In-App Purchase that the system records in the transaction.
type price int64

// The number of purchased consumable products.
type quantity int32

// The reason for a refunded transaction.
// 0: The App Store refunded the transaction on behalf of the customer for other reasons, for example, an accidental purchase.
// 1: The App Store refunded the transaction on behalf of the customer due to an actual or perceived issue within your app.
type revocationReason int32

// The three-letter code that represents the country or region associated with the App Store storefront of the purchase.
type storefront string

// An Apple-defined value that uniquely identifies an App Store storefront.
type storefrontId string

// The cause of a purchase transaction, which indicates whether it’s a customer’s purchase or a renewal for an auto-renewable subscription that the system initiates.
// PURCHASE: The customer initiated the purchase, which may be for any in-app purchase type: consumable, non-consumable, non-renewing subscription, or auto-renewable subscription.
// RENEWAL: The App Store server initiated the purchase transaction to renew an auto-renewable subscription.
type transactionReason string

type JWSTransactionDecodedPayload struct {
	// A UUID you create at the time of purchase that associates the transaction with a customer on your own service.
	// If your app doesn’t provide an appAccountToken, this string is empty. For more information, see appAccountToken(_:).
	AppAccountToken UUID `json:"appAccountToken"`

	// The bundle identifier of the app.
	BundleId BundleId `json:"bundleId"`

	// The three-letter ISO 4217 currency code associated with the price parameter. This value is present only if price is present.
	Currency currency `json:"currency"`

	// The server environment, either sandbox or production.
	Environment Environment `json:"environment"`

	// The UNIX time, in milliseconds, that the subscription expires or renews.
	ExpiresDate Timestamp `json:"expiresDate"`

	// A string that describes whether the transaction was purchased by the customer, or is available to them through Family Sharing.
	InAppOwnershipType InAppOwnershipType `json:"inAppOwnershipType"`

	// A Boolean value that indicates whether the customer upgraded to another subscription.
	IsUpgraded isUpgraded `json:"isUpgraded"`

	// The payment mode you configure for the subscription offer, such as Free Trial, Pay As You Go, or Pay Up Front.
	OfferDiscountType offerDiscountType `json:"offerDiscountType"`

	// The identifier that contains the offer code or the promotional offer identifier.
	OfferIdentifier offerIdentifier `json:"offerIdentifier"`

	// A value that represents the promotional offer type.
	OfferType offerType `json:"offerType"`

	// The UNIX time, in milliseconds, that represents the purchase date of the original transaction identifier.
	OriginalPurchaseDate Timestamp `json:"originalPurchaseDate"`

	// The transaction identifier of the original purchase.
	OriginalTransactionId OriginalTransactionId `json:"originalTransactionId"`

	// An integer value that represents the price multiplied by 1000 of the in-app purchase or subscription offer
	// you configured in App Store AppStoreConnectAPI and that the system records at the time of the purchase. For more information,
	// see price. The currency parameter indicates the currency of this price.
	Price price `json:"price"`

	// The unique identifier of the product.
	ProductId ProductId `json:"productId"`

	// The UNIX time, in milliseconds, that the App Store charged the customer’s account for a purchase, restored product,
	// subscription, or subscription renewal after a lapse.
	PurchaseDate Timestamp `json:"purchaseDate"`

	// The number of consumable products the customer purchased.
	Quantity quantity `json:"quantity"`

	// The UNIX time, in milliseconds, that the App Store refunded the transaction or revoked it from Family Sharing.
	RevocationDate Timestamp `json:"revocationDate"`

	// The reason that the App Store refunded the transaction or revoked it from Family Sharing.
	RevocationReason revocationReason `json:"revocationReason"`

	// The UNIX time, in milliseconds, that the App Store signed the JSON Web Signature (JWS) data.
	SignedDate Timestamp `json:"signedDate"`

	// The three-letter code that represents the country or region associated with the App Store storefront for the purchase.
	Storefront storefront `json:"storefront"`

	// An Apple-defined value that uniquely identifies the App Store storefront associated with the purchase.
	StorefrontId storefrontId `json:"storefrontId"`

	// The identifier of the subscription group to which the subscription belongs.
	SubscriptionGroupIdentifier SubscriptionGroupIdentifier `json:"subscriptionGroupIdentifier"`

	// The unique identifier of the transaction.
	TransactionId TransactionId `json:"transactionId"`

	// The reason for the purchase transaction, which indicates whether it’s a customer’s purchase or a renewal for an auto-renewable subscription that the system initates.
	TransactionReason transactionReason `json:"transactionReason"`

	// The type of the in-app purchase.
	Type string `json:"type"`

	// The unique identifier of subscription purchase events across devices, including subscription renewals.
	WebOrderLineItemId WebOrderLineItemId `json:"webOrderLineItemId"`
}

// JWSTransaction Transaction information signed by the App Store, in JSON Web Signature (JWS) Compact Serialization format.
type JWSTransaction string

// Decrypt 解密数据
func (j JWSTransaction) Decrypt() (*JWSTransactionDecodedPayload, error) {
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

	var payload JWSTransactionDecodedPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload JSON: %v", err)
	}

	singPayload := string(j)
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

// Parse signedPayload and return Header, Payload and Signature
func (j JWSTransaction) parseSignedPayload() (*JWSDecodedHeader, []byte, []byte, error) {
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
