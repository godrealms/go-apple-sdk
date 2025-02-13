package types

// ExternalPurchaseId The field of an external purchase token that uniquely identifies the token.
type ExternalPurchaseId string

func (e ExternalPurchaseId) String() string {
	return string(e)
}
