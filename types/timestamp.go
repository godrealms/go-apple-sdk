package types

import "time"

type Timestamp int64

func (t Timestamp) Time() time.Time {
	// Convert milliseconds to seconds for Unix timestamp
	seconds := int64(t) / 1000
	// Convert to Go's time.Time
	return time.Unix(seconds, 0).UTC()
}
