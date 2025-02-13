package types

// UserStatus
// The status of a customer’s account within your app.
// 0: Account status is undeclared. Use this value to avoid providing information for this field.
// 1: The customer’s account is active.
// 2: The customer’s account is suspended.
// 3: The customer’s account is terminated.
// 4: The customer’s account has limited access.
type UserStatus int32
