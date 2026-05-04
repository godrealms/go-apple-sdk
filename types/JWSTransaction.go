package types

import "github.com/godrealms/go-apple-sdk/jws"

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

// JWSTransaction is the JWS-Compact-Serialised transaction Apple
// returns from App Store Server API endpoints. Decrypt verifies
// and decodes it using the package-default Verifier (Apple Root
// CA G3 + Apple receipt-signing OID required). To override the
// trust anchors (for tests or future Apple cert rotation), use
// DecryptWith with a custom *jws.Verifier.
type JWSTransaction string

// Decrypt verifies the JWS chain + signature and returns the
// decoded payload. Returns *jws.VerificationError on failure;
// inspect VerificationError.Reason to distinguish chain failure
// from OID mismatch from signature mismatch from malformed input.
func (j JWSTransaction) Decrypt() (*JWSTransactionDecodedPayload, error) {
	return jws.VerifyAndDecode[JWSTransactionDecodedPayload](jws.DefaultVerifier(), string(j))
}

// DecryptWith verifies using the supplied Verifier instead of the
// package default. Use this when you need a custom root CA pool
// (e.g. integration tests that mint their own chain).
func (j JWSTransaction) DecryptWith(v *jws.Verifier) (*JWSTransactionDecodedPayload, error) {
	return jws.VerifyAndDecode[JWSTransactionDecodedPayload](v, string(j))
}
