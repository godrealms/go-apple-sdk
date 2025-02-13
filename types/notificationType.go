package types

// NotificationType represents the main type of an App Store server notification.
type NotificationType string

const (
	// NOTIFICATION_TYPE_CONSUMPTION_REQUEST
	// Indicates that the App Store is requesting consumption data for a refund request initiated by the customer.
	// For more information, see Send Consumption Information.
	NOTIFICATION_TYPE_CONSUMPTION_REQUEST NotificationType = "CONSUMPTION_REQUEST"

	// NOTIFICATION_TYPE_DID_CHANGE_RENEWAL_PREF
	// Indicates that the customer made a change to their subscription plan.
	// If the subtype is UPGRADE, the user upgraded their subscription (effective immediately).
	// If the subtype is DOWNGRADE, the customer downgraded their subscription (effective at the next renewal date).
	NOTIFICATION_TYPE_DID_CHANGE_RENEWAL_PREF NotificationType = "DID_CHANGE_RENEWAL_PREF"

	// NOTIFICATION_TYPE_DID_CHANGE_RENEWAL_STATUS
	// Indicates that the customer made a change to the subscription renewal status.
	// Subtypes include AUTO_RENEW_ENABLED and AUTO_RENEW_DISABLED.
	NOTIFICATION_TYPE_DID_CHANGE_RENEWAL_STATUS NotificationType = "DID_CHANGE_RENEWAL_STATUS"

	// NOTIFICATION_TYPE_DID_FAIL_TO_RENEW
	// Indicates that the subscription failed to renew due to a billing issue.
	// Subtypes include GRACE_PERIOD (if the subscription is in a grace period) or empty (if not in a grace period).
	NOTIFICATION_TYPE_DID_FAIL_TO_RENEW NotificationType = "DID_FAIL_TO_RENEW"

	// NOTIFICATION_TYPE_DID_RENEW
	// Indicates that the subscription successfully renewed.
	// Subtypes include BILLING_RECOVERY (if the subscription was previously expired and has been restored) or empty.
	NOTIFICATION_TYPE_DID_RENEW NotificationType = "DID_RENEW"

	// NOTIFICATION_TYPE_EXPIRED
	// Indicates that a subscription expired.
	// Subtypes include VOLUNTARY, BILLING_RETRY, PRICE_INCREASE, and PRODUCT_NOT_FOR_SALE.
	NOTIFICATION_TYPE_EXPIRED NotificationType = "EXPIRED"

	// NOTIFICATION_TYPE_GRACE_PERIOD_EXPIRED
	// Indicates that the billing grace period has ended without renewing the subscription.
	NOTIFICATION_TYPE_GRACE_PERIOD_EXPIRED NotificationType = "GRACE_PERIOD_EXPIRED"

	// NOTIFICATION_TYPE_OFFER_REDEEMED
	// Indicates that a customer with an active subscription redeemed a subscription offer.
	// Subtypes include UPGRADE (immediate upgrade), DOWNGRADE (effective at next renewal date), or empty.
	NOTIFICATION_TYPE_OFFER_REDEEMED NotificationType = "OFFER_REDEEMED"

	// NOTIFICATION_TYPE_ONE_TIME_CHARGE
	// Indicates that the customer purchased a consumable, non-consumable, or non-renewing subscription.
	// This notification is currently available only in the sandbox environment.
	NOTIFICATION_TYPE_ONE_TIME_CHARGE NotificationType = "ONE_TIME_CHARGE"

	// NOTIFICATION_TYPE_PRICE_INCREASE
	// Indicates that the system informed the customer of a price increase for an auto-renewable subscription.
	// Subtypes include PENDING (customer hasn’t responded) or ACCEPTED (customer consented to the price increase).
	NOTIFICATION_TYPE_PRICE_INCREASE NotificationType = "PRICE_INCREASE"

	// NOTIFICATION_TYPE_REFUND
	// Indicates that the App Store successfully refunded a transaction for a consumable, non-consumable,
	// auto-renewable subscription, or non-renewing subscription.
	NOTIFICATION_TYPE_REFUND NotificationType = "REFUND"

	// NOTIFICATION_TYPE_REFUND_DECLINED
	// Indicates that the App Store declined a refund request.
	NOTIFICATION_TYPE_REFUND_DECLINED NotificationType = "REFUND_DECLINED"

	// NOTIFICATION_TYPE_REFUND_REVERSED
	// Indicates that the App Store reversed a previously granted refund due to a dispute raised by the customer.
	NOTIFICATION_TYPE_REFUND_REVERSED NotificationType = "REFUND_REVERSED"

	// NOTIFICATION_TYPE_RENEWAL_EXTENDED
	// Indicates that the App Store extended the subscription renewal date for a specific subscription.
	NOTIFICATION_TYPE_RENEWAL_EXTENDED NotificationType = "RENEWAL_EXTENDED"

	// NOTIFICATION_TYPE_RENEWAL_EXTENSION
	// Indicates that the App Store attempted to extend the subscription renewal date.
	// Subtypes include SUMMARY (extension completed for all eligible subscribers) or FAILURE (extension failed for a specific subscription).
	NOTIFICATION_TYPE_RENEWAL_EXTENSION NotificationType = "RENEWAL_EXTENSION"

	// NOTIFICATION_TYPE_REVOKE
	// Indicates that an in-app purchase entitlement through Family Sharing is no longer available.
	NOTIFICATION_TYPE_REVOKE NotificationType = "REVOKE"

	// NOTIFICATION_TYPE_SUBSCRIBED
	// Indicates that the customer subscribed to an auto-renewable subscription.
	// Subtypes include INITIAL_BUY (first-time purchase) or RESUBSCRIBE (resubscribed to the same or another subscription).
	NOTIFICATION_TYPE_SUBSCRIBED NotificationType = "SUBSCRIBED"

	// NOTIFICATION_TYPE_TEST
	// Indicates a test notification sent by the App Store server when you request it.
	NOTIFICATION_TYPE_TEST NotificationType = "TEST"
)
