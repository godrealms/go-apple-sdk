package types

import (
	"testing"
	"time"
)

// TestTimestamp_Time_PreservesMillis guards against the millisecond-truncation
// bug: Apple timestamps are Unix milliseconds, so Time() must not divide by
// 1000 and drop the sub-second part.
func TestTimestamp_Time_PreservesMillis(t *testing.T) {
	cases := []struct {
		name string
		in   Timestamp
		want time.Time
	}{
		{"epoch", 0, time.UnixMilli(0).UTC()},
		{"whole_second", 1700000000000, time.UnixMilli(1700000000000).UTC()},
		{"with_millis", 1700000000123, time.UnixMilli(1700000000123).UTC()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.Time()
			if !got.Equal(tc.want) {
				t.Errorf("Time() = %s, want %s",
					got.Format(time.RFC3339Nano), tc.want.Format(time.RFC3339Nano))
			}
		})
	}
}
