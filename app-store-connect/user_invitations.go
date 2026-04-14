package AppStoreConnect

import (
	"context"
	"time"
)

// UserInvitationsService provides access to /v1/userInvitations, the
// pending-invitation catalog for new team members.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/user_invitations
type UserInvitationsService struct {
	svc *Service
}

// UserInvitation is a typed alias for a JSON:API userInvitations resource.
type UserInvitation = Resource[UserInvitationAttributes]

// UserInvitationAttributes models the attributes of a userInvitations resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/userinvitation/attributes
type UserInvitationAttributes struct {
	Email               string     `json:"email,omitempty"`
	FirstName           string     `json:"firstName,omitempty"`
	LastName            string     `json:"lastName,omitempty"`
	ExpirationDate      *time.Time `json:"expirationDate,omitempty"`
	Roles               []string   `json:"roles,omitempty"`
	AllAppsVisible      *bool      `json:"allAppsVisible,omitempty"`
	ProvisioningAllowed *bool      `json:"provisioningAllowed,omitempty"`
}

// ListUserInvitationsResponse is the decoded response for [UserInvitationsService.List].
type ListUserInvitationsResponse struct {
	Data     []UserInvitation `json:"data"`
	Included []Resource[any]  `json:"included,omitempty"`
	Links    *Links           `json:"links,omitempty"`
}

// List returns a page of pending user invitations.
func (s *UserInvitationsService) List(ctx context.Context, query *Query) (*ListUserInvitationsResponse, error) {
	var doc Document[UserInvitationAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/userInvitations", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListUserInvitationsResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListIterator returns a paginator that walks every page of user invitations.
func (s *UserInvitationsService) ListIterator(query *Query) *Paginator[UserInvitationAttributes] {
	return newPaginator[UserInvitationAttributes](s.svc, "/v1/userInvitations", query)
}

// CreateUserInvitationRequest describes a team invitation.
// Email, FirstName, LastName, and Roles are all required.
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/create_a_user_invitation
type CreateUserInvitationRequest struct {
	Email               string
	FirstName           string
	LastName            string
	Roles               []string
	AllAppsVisible      *bool
	ProvisioningAllowed *bool
	VisibleAppIDs       []string // if AllAppsVisible is false, must list explicit app ids
}

// Create sends a team invitation to the given email.
func (s *UserInvitationsService) Create(ctx context.Context, req CreateUserInvitationRequest) (*UserInvitation, error) {
	if req.Email == "" || req.FirstName == "" || req.LastName == "" {
		return nil, &ClientError{Message: "UserInvitations.Create: Email, FirstName, and LastName are required"}
	}
	if len(req.Roles) == 0 {
		return nil, &ClientError{Message: "UserInvitations.Create: at least one role is required"}
	}
	attrs := map[string]any{
		"email":     req.Email,
		"firstName": req.FirstName,
		"lastName":  req.LastName,
		"roles":     req.Roles,
	}
	if req.AllAppsVisible != nil {
		attrs["allAppsVisible"] = *req.AllAppsVisible
	}
	if req.ProvisioningAllowed != nil {
		attrs["provisioningAllowed"] = *req.ProvisioningAllowed
	}
	data := map[string]any{
		"type":       "userInvitations",
		"attributes": attrs,
	}
	if len(req.VisibleAppIDs) > 0 {
		data["relationships"] = map[string]any{
			"visibleApps": map[string]any{
				"data": buildIdentifiers("apps", req.VisibleAppIDs),
			},
		}
	}
	body := map[string]any{"data": data}
	var doc Document[UserInvitationAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/userInvitations", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete cancels a pending invitation.
func (s *UserInvitationsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "UserInvitations.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/userInvitations/"+id, nil, nil, nil)
	return err
}
