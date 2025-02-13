package types

import "strconv"

// SucceededCount The count of subscriptions that successfully receive a subscription-renewal-date extension.
type SucceededCount int64

func (s SucceededCount) String() string {
	return strconv.FormatInt(int64(s), 10)
}

func (s SucceededCount) Int64() int64 {
	return int64(s)
}
