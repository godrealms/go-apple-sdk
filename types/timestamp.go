package types

import "time"

type Timestamp int64

func (t Timestamp) Time() time.Time {
	// Apple timestamps are Unix milliseconds; preserve the sub-second part
	// instead of truncating to whole seconds.
	return time.UnixMilli(int64(t)).UTC()
}
