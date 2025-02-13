package types

// AccountTenure
// The age of the customer’s account.
// 0: Account age is undeclared. Use this value to avoid providing information for this field.
// 1: Account age is between 0–3 days.
// 2: Account age is between 3–10 days.
// 3: Account age is between 10–30 days.
// 4: Account age is between 30–90 days.
// 5: Account age is between 90–180 days.
// 6: Account age is between 180–365 days.
// 7: Account age is over 365 days.
type AccountTenure int32
