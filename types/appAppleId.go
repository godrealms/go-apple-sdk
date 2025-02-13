package types

// AppAppleId The unique identifier of an app in the App Store.
type AppAppleId int64

func (a AppAppleId) Int64() int64 {
	return int64(a)
}
