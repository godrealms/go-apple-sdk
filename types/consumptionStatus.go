package types

// ConsumptionStatus
// A value that indicates the extent to which the customer consumed the in-app purchase.
// 0: The consumption status is undeclared. Use this value to avoid providing information for this field.
// 1: The in-app purchase is not consumed.
// 2: The in-app purchase is partially consumed.
// 3: The in-app purchase is fully consumed.
type ConsumptionStatus int32
