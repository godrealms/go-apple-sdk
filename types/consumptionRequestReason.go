package types

// ConsumptionRequestReason The customer-provided reason for a refund request.
type ConsumptionRequestReason string

const (
	ConsumptionRequestReasonUnintendedPurchase      ConsumptionRequestReason = "UNINTENDED_PURCHASE"       // The customer didn’t intend to make the in-app purchase.
	ConsumptionRequestReasonFulfillmentIssue        ConsumptionRequestReason = "FULFILLMENT_ISSUE"         // The customer had issues with receiving or using the in-app purchase.
	ConsumptionRequestReasonUnsatisfiedWithPurchase ConsumptionRequestReason = "UNSATISFIED_WITH_PURCHASE" // The customer wasn’t satisfied with the in-app purchase.
	ConsumptionRequestReasonLegal                   ConsumptionRequestReason = "LEGAL"                     // The customer requested a refund based on a legal reason.
	ConsumptionRequestReasonOther                   ConsumptionRequestReason = "OTHER"                     // The customer requested a refund for other reasons.
)

func (r ConsumptionRequestReason) String() string {
	return string(r)
}
