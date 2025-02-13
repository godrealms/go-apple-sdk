package types

// LifetimeDollarsRefunded
// A value that indicates the dollar amount of refunds the customer has received in your app, since purchasing the app, across all platforms.
// 0: Lifetime refund amount is undeclared. Use this value to avoid providing information for this field.
// 1: Lifetime refund amount is 0 USD.
// 2: Lifetime refund amount is between 0.01–49.99 USD.
// 3: Lifetime refund amount is between 50–99.99 USD.
// 4: Lifetime refund amount is between 100–499.99 USD.
// 5: Lifetime refund amount is between 500–999.99 USD.
// 6: Lifetime refund amount is between 1000–1999.99 USD.
// 7: Lifetime refund amount is over 2000 USD.
type LifetimeDollarsRefunded int32
