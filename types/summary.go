package types

type Summary struct {
	// The UUID that represents a specific request to extend a subscription renewal date.
	// This value matches the value you initially specify in the requestIdentifier when you call Extend Subscription
	// Renewal Dates for All Active Subscribers in the App Store AppStoreServerAPI API.
	RequestIdentifier RequestIdentifier `json:"requestIdentifier"`

	// The server environment that the notification applies to, either sandbox or production.
	Environment Environment `json:"environment"`

	// The unique identifier of the app that the notification applies to. This property is available for apps that users
	// download from the App Store. It isn’t present in the sandbox environment.
	AppAppleId AppAppleId `json:"appAppleId"`

	// The bundle identifier of the app.
	BundleId BundleId `json:"bundleId"`

	// The product identifier of the auto-renewable subscription that the subscription-renewal-date extension applies to.
	ProductId ProductId `json:"productId"`

	// A list of country codes that limits the App Store’s attempt to apply the subscription-renewal-date extension.
	// If this list isn’t present, the subscription-renewal-date extension applies to all storefronts.
	StorefrontCountryCodes StorefrontCountryCodes `json:"storefrontCountryCodes"`

	// The final count of subscriptions that fail to receive a subscription-renewal-date extension.
	FailedCount FailedCount `json:"failedCount"`

	// The final count of subscriptions that successfully receive a subscription-renewal-date extension.
	SucceededCount SucceededCount `json:"succeededCount"`
}
