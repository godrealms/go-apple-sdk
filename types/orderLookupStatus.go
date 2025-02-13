package types

// OrderLookupStatus A value that indicates whether the order ID in the request is valid for your app.
// 0: The orderId that you provided in the Look Up Order ID request is valid and contains at least one in-app purchase for your app.
// 1: The orderId is invalid or doesn’t contain any in-app purchases for your app.
type OrderLookupStatus int32
