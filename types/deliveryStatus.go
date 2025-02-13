package types

// DeliveryStatus
// A value that indicates whether the app successfully delivered an in-app purchase that works properly.
// 0: The app delivered the consumable in-app purchase and it’s working properly.
// 1: The app didn’t deliver the consumable in-app purchase due to a quality issue.
// 2: The app delivered the wrong item.
// 3: The app didn’t deliver the consumable in-app purchase due to a server outage.
// 4: The app didn’t deliver the consumable in-app purchase due to an in-game currency change.
// 5: The app didn’t deliver the consumable in-app purchase for other reasons.
type DeliveryStatus int32
