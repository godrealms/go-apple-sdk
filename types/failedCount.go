package types

import "strconv"

// FailedCount The count of subscriptions that fail to receive a subscription-renewal-date extension.
type FailedCount int64

func (f FailedCount) String() string {
	return strconv.FormatInt(int64(f), 10)
}

func (f FailedCount) Int64() int64 {
	return int64(f)
}
