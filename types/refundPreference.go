package types

// RefundPreference
// A value that indicates your preferred outcome for the refund request.
// 0: The refund preference is undeclared. Use this value to avoid providing information for this field.
// 1: You prefer that Apple grants the refund.
// 2: You prefer that Apple declines the refund.
// 3: You have no preference whether Apple grants or declines the refund.
type RefundPreference int32
