package types

// LifetimeDollarsPurchased
// A value that indicates the dollar amount of in-app purchases the customer has made in your app, since purchasing the app, across all platforms.
// 0: Lifetime purchase amount is undeclared. Use this value to avoid providing information for this field.
// 1: Lifetime purchase amount is 0 USD.
// 2: Lifetime purchase amount is between 0.01–49.99 USD.
// 3: Lifetime purchase amount is between 50–99.99 USD.
// 4: Lifetime purchase amount is between 100–499.99 USD.
// 5: Lifetime purchase amount is between 500–999.99 USD.
// 6: Lifetime purchase amount is between 1000–1999.99 USD.
// 7: Lifetime purchase amount is over 2000 USD.
type LifetimeDollarsPurchased int32
