package AppStoreConnect

import "context"

// UsersService provides access to /v1/users, the team member catalog.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/users
type UsersService struct {
	svc *Service
}

// User is a typed alias for a JSON:API users resource.
type User = Resource[UserAttributes]

// UserAttributes models the attributes of a users resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/user/attributes
type UserAttributes struct {
	Username            string   `json:"username,omitempty"`
	FirstName           string   `json:"firstName,omitempty"`
	LastName            string   `json:"lastName,omitempty"`
	Roles               []string `json:"roles,omitempty"`
	AllAppsVisible      *bool    `json:"allAppsVisible,omitempty"`
	ProvisioningAllowed *bool    `json:"provisioningAllowed,omitempty"`
}

// ListUsersResponse is the decoded response for [UsersService.List].
type ListUsersResponse struct {
	Data     []User          `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// List returns a page of users matching the query.
func (s *UsersService) List(ctx context.Context, query *Query) (*ListUsersResponse, error) {
	var doc Document[UserAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/users", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListUsersResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListIterator returns a paginator that walks every page of users.
func (s *UsersService) ListIterator(query *Query) *Paginator[UserAttributes] {
	return newPaginator[UserAttributes](s.svc, "/v1/users", query)
}

// Get fetches a single user by resource id.
func (s *UsersService) Get(ctx context.Context, id string, query *Query) (*User, error) {
	if id == "" {
		return nil, &ClientError{Message: "Users.Get: id is required"}
	}
	var doc Document[UserAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/users/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// UserUpdate collects the mutable fields on a user. Only the fields
// set via setters are sent.
type UserUpdate struct {
	attrs map[string]any
}

// NewUserUpdate returns an empty [UserUpdate].
func NewUserUpdate() *UserUpdate { return &UserUpdate{attrs: make(map[string]any)} }

// Roles sets the user's roles (e.g. "DEVELOPER", "APP_MANAGER").
func (u *UserUpdate) Roles(roles ...string) *UserUpdate {
	u.attrs["roles"] = roles
	return u
}

// AllAppsVisible toggles whether the user can see every app in the team.
func (u *UserUpdate) AllAppsVisible(v bool) *UserUpdate {
	u.attrs["allAppsVisible"] = v
	return u
}

// ProvisioningAllowed toggles whether the user can manage certificates
// and provisioning profiles.
func (u *UserUpdate) ProvisioningAllowed(v bool) *UserUpdate {
	u.attrs["provisioningAllowed"] = v
	return u
}

// IsEmpty reports whether any attribute change is pending.
func (u *UserUpdate) IsEmpty() bool { return len(u.attrs) == 0 }

// Update modifies the given user's roles and visibility.
func (s *UsersService) Update(ctx context.Context, id string, update *UserUpdate) (*User, error) {
	if id == "" {
		return nil, &ClientError{Message: "Users.Update: id is required"}
	}
	if update == nil || update.IsEmpty() {
		return nil, &ClientError{Message: "Users.Update: update has no attribute changes"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type":       "users",
			"id":         id,
			"attributes": update.attrs,
		},
	}
	var doc Document[UserAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/users/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a user from the team.
func (s *UsersService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "Users.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/users/"+id, nil, nil, nil)
	return err
}
