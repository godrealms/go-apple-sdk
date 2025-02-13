package types

type Status int32

const (
	StatusActive      Status = 1 //The auto-renewable subscription is active.
	StatusExpired     Status = 2 //The auto-renewable subscription is expired.
	StatusRetryPeriod Status = 3 //The auto-renewable subscription is in a billing retry period.
	StatusGracePeriod Status = 4 //The auto-renewable subscription is in a Billing Grace Period.
	StatusRevoked     Status = 5 //The auto-renewable subscription is revoked.
)
