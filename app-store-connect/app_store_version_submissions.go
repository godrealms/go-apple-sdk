package AppStoreConnect

import "context"

// AppStoreVersionSubmissionsService provides access to
// /v1/appStoreVersionSubmissions, the endpoint that actually submits
// an App Store version to Apple's review queue.
//
// Creating a submission for a version is the programmatic equivalent
// of pressing "Submit for Review" in App Store Connect.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/app_store_version_submissions
type AppStoreVersionSubmissionsService struct {
	svc *Service
}

// AppStoreVersionSubmission is a typed alias for a JSON:API
// appStoreVersionSubmissions resource. The resource is
// relationship-only — there are no interesting attributes to model.
type AppStoreVersionSubmission = Resource[struct{}]

// Submit submits the given App Store version for review. On success
// Apple returns the created submission resource; calling Submit a
// second time for the same version will fail with STATE_ERROR.
func (s *AppStoreVersionSubmissionsService) Submit(ctx context.Context, versionID string) (*AppStoreVersionSubmission, error) {
	if versionID == "" {
		return nil, &ClientError{Message: "AppStoreVersionSubmissions.Submit: versionID is required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "appStoreVersionSubmissions",
			"relationships": map[string]any{
				"appStoreVersion": map[string]any{
					"data": map[string]any{"type": "appStoreVersions", "id": versionID},
				},
			},
		},
	}
	var doc Document[struct{}]
	if _, err := s.svc.do(ctx, "POST", "/v1/appStoreVersionSubmissions", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Cancel withdraws a pending submission by its resource id.
// Apple rejects cancellation after review has started.
func (s *AppStoreVersionSubmissionsService) Cancel(ctx context.Context, submissionID string) error {
	if submissionID == "" {
		return &ClientError{Message: "AppStoreVersionSubmissions.Cancel: submissionID is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/appStoreVersionSubmissions/"+submissionID, nil, nil, nil)
	return err
}
