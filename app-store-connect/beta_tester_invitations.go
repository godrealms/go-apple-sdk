package AppStoreConnect

import "context"

// BetaTesterInvitationsService sends TestFlight invitations to an
// existing beta tester for a given app. This is the trigger that
// actually emails the tester — creating a beta tester resource
// without this call only records them in the account.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/beta_tester_invitations
type BetaTesterInvitationsService struct {
	svc *Service
}

// BetaTesterInvitation is a typed alias for a JSON:API
// betaTesterInvitations resource. The resource has no useful
// attributes — it is essentially a relationship-only record — so the
// attribute type is a blank struct.
type BetaTesterInvitation = Resource[struct{}]

// Create sends an invitation email to the given beta tester for the
// given app. Both IDs are required.
func (s *BetaTesterInvitationsService) Create(ctx context.Context, appID, betaTesterID string) (*BetaTesterInvitation, error) {
	if appID == "" {
		return nil, &ClientError{Message: "BetaTesterInvitations.Create: appID is required"}
	}
	if betaTesterID == "" {
		return nil, &ClientError{Message: "BetaTesterInvitations.Create: betaTesterID is required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "betaTesterInvitations",
			"relationships": map[string]any{
				"app": map[string]any{
					"data": map[string]any{"type": "apps", "id": appID},
				},
				"betaTester": map[string]any{
					"data": map[string]any{"type": "betaTesters", "id": betaTesterID},
				},
			},
		},
	}
	var doc Document[struct{}]
	if _, err := s.svc.do(ctx, "POST", "/v1/betaTesterInvitations", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}
