package types

// RequestIdentifier A string that contains a unique identifier for a subscription-renewal-date extension request.
type RequestIdentifier string // UUID

func (r RequestIdentifier) String() string {
	return string(r)
}
