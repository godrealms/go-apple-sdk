package types

// PlayTime
// A value that indicates the amount of time that the customer used the app.
// 0: The engagement time is undeclared. Use this value to avoid providing information for this field.
// 1: The engagement time is between 0–5 minutes.
// 2: The engagement time is between 5–60 minutes.
// 3: The engagement time is between 1–6 hours.
// 4: The engagement time is between 6–24 hours.
// 5: The engagement time is between 1–4 days.
// 6: The engagement time is between 4–16 days.
// 7: The engagement time is over 16 days.
type PlayTime int32
