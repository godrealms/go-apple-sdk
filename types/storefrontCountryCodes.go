package types

// StorefrontCountryCode The three-letter code that represents the country or region associated with the App Store storefront.
type StorefrontCountryCode string // This type uses the ISO 3166-1 Alpha-3 country code representation.

// IsValid reports whether s is a well-formed ISO 3166-1 Alpha-3 storefront
// code, i.e. exactly three characters.
func (s StorefrontCountryCode) IsValid() bool {
	return len(s) == 3
}

func (s StorefrontCountryCode) String() string {
	return string(s)
}

// StorefrontCountryCodes A list of storefront country codes for limiting the storefronts for a subscription-renewal-date extension.
type StorefrontCountryCodes []StorefrontCountryCode // Maximum length: 3
