package types

// Subtype represents additional information about a notification type.
type Subtype string

const (
	// SUBTYPE_UPGRADE
	// Indicates that the user upgraded their subscription plan.
	// The upgrade goes into effect immediately, starting a new billing period.
	// The user receives a prorated refund for the unused portion of the previous period.
	SUBTYPE_UPGRADE Subtype = "UPGRADE"

	// SUBTYPE_DOWNGRADE
	// Indicates that the user downgraded their subscription plan.
	// The downgrade takes effect at the next renewal date and doesn’t affect the currently active plan.
	SUBTYPE_DOWNGRADE Subtype = "DOWNGRADE"

	// SUBTYPE_AUTO_RENEW_ENABLED
	// Indicates that the customer reenabled subscription auto-renewal.
	SUBTYPE_AUTO_RENEW_ENABLED Subtype = "AUTO_RENEW_ENABLED"

	// SUBTYPE_AUTO_RENEW_DISABLED
	// Indicates that the customer disabled subscription auto-renewal,
	// or the App Store disabled subscription auto-renewal after the customer requested a refund.
	SUBTYPE_AUTO_RENEW_DISABLED Subtype = "AUTO_RENEW_DISABLED"

	// SUBTYPE_BILLING_RECOVERY
	// Indicates that an expired subscription that previously failed to renew has successfully renewed.
	SUBTYPE_BILLING_RECOVERY Subtype = "BILLING_RECOVERY"

	// SUBTYPE_GRACE_PERIOD
	// Indicates that the subscription is in a billing grace period.
	SUBTYPE_GRACE_PERIOD Subtype = "GRACE_PERIOD"

	// SUBTYPE_PENDING
	// Indicates that the customer hasn’t responded to a subscription price increase that requires customer consent.
	SUBTYPE_PENDING Subtype = "PENDING"

	// SUBTYPE_ACCEPTED
	// Indicates that the customer consented to a subscription price increase.
	SUBTYPE_ACCEPTED Subtype = "ACCEPTED"

	// SUBTYPE_VOLUNTARY
	// Indicates that the subscription expired after the user disabled subscription renewal.
	SUBTYPE_VOLUNTARY Subtype = "VOLUNTARY"

	// SUBTYPE_BILLING_RETRY
	// Indicates that the subscription expired because the billing retry period ended without a successful billing transaction.
	SUBTYPE_BILLING_RETRY Subtype = "BILLING_RETRY"

	// SUBTYPE_PRICE_INCREASE
	// Indicates that the subscription expired because the customer didn’t consent to a price increase that requires customer consent.
	SUBTYPE_PRICE_INCREASE Subtype = "PRICE_INCREASE"

	// SUBTYPE_PRODUCT_NOT_FOR_SALE
	// Indicates that the subscription expired because the product wasn’t available for purchase at the time the subscription attempted to renew.
	SUBTYPE_PRODUCT_NOT_FOR_SALE Subtype = "PRODUCT_NOT_FOR_SALE"

	// SUBTYPE_INITIAL_BUY
	// Indicates that the customer either purchased or received access through Family Sharing to the subscription for the first time.
	SUBTYPE_INITIAL_BUY Subtype = "INITIAL_BUY"

	// SUBTYPE_RESUBSCRIBE
	// Indicates that the user resubscribed or received access through Family Sharing to the same subscription or another subscription within the same subscription group.
	SUBTYPE_RESUBSCRIBE Subtype = "RESUBSCRIBE"

	// SUBTYPE_SUMMARY
	// Indicates that the App Store completed extending the renewal date for all eligible subscribers.
	SUBTYPE_SUMMARY Subtype = "SUMMARY"

	// SUBTYPE_FAILURE
	// Indicates that the App Store didn’t succeed in extending the renewal date for a specific subscription.
	SUBTYPE_FAILURE Subtype = "FAILURE"

	// SUBTYPE_UNREPORTED
	// Indicates that Apple created an external purchase token for your app but didn’t receive a report.
	SUBTYPE_UNREPORTED Subtype = "UNREPORTED"
)
