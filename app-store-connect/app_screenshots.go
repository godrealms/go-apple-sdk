package AppStoreConnect

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
)

// AppScreenshotsService manages /v1/appScreenshots — Apple's per-image
// resources under an [AppScreenshotSet]. Uploading an image is a
// three-step flow:
//
//  1. Reserve: POST /v1/appScreenshots with the intended filename +
//     byte length. Apple responds with one or more "uploadOperations"
//     telling the client exactly which URL to PUT to, at what offset,
//     with which headers.
//  2. Execute: iterate every upload operation and PUT the relevant
//     byte slice of the source file. Apple streams to Amazon S3, so
//     these requests do NOT carry App Store Connect auth.
//  3. Commit: PATCH /v1/appScreenshots/{id} with uploaded=true and
//     an MD5 checksum of the full source file.
//
// Most callers want [AppScreenshotsService.Upload], which performs
// all three steps inline from an in-memory byte slice. The finer-grained
// [Reserve], [ExecuteUploadOperations], and [Commit] entry points are
// exported for callers that stream from disk or retry individual steps.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/uploading_assets_to_app_store_connect
type AppScreenshotsService struct {
	svc *Service
}

// AppScreenshot is a typed alias for a JSON:API appScreenshots resource.
type AppScreenshot = Resource[AppScreenshotAttributes]

// AppScreenshotAttributes models an appScreenshots resource. The
// UploadOperations array is only populated on the reservation response.
type AppScreenshotAttributes struct {
	FileSize           int64            `json:"fileSize,omitempty"`
	FileName           string           `json:"fileName,omitempty"`
	SourceFileChecksum string           `json:"sourceFileChecksum,omitempty"`
	ImageAsset         *ImageAsset      `json:"imageAsset,omitempty"`
	AssetToken         string           `json:"assetToken,omitempty"`
	AssetDeliveryState *struct {
		State    string `json:"state,omitempty"`
		Warnings []struct {
			Code        string `json:"code,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"warnings,omitempty"`
		Errors []struct {
			Code        string `json:"code,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"errors,omitempty"`
	} `json:"assetDeliveryState,omitempty"`
	Uploaded         *bool             `json:"uploaded,omitempty"`
	UploadOperations []UploadOperation `json:"uploadOperations,omitempty"`
}

// ImageAsset is the delivered asset reference that Apple returns after
// processing completes.
type ImageAsset struct {
	TemplateURL string `json:"templateUrl,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// UploadOperation describes one byte range to PUT to Apple's upload
// endpoint. Apple may split a single screenshot across several
// operations; each request must include every header in Headers.
type UploadOperation struct {
	Method  string                 `json:"method,omitempty"`
	URL     string                 `json:"url,omitempty"`
	Length  int64                  `json:"length,omitempty"`
	Offset  int64                  `json:"offset,omitempty"`
	Headers []UploadOperationHeader `json:"requestHeaders,omitempty"`
}

// UploadOperationHeader is one name/value pair Apple wants on the PUT.
type UploadOperationHeader struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// ReserveAppScreenshotRequest describes the reservation metadata. All
// three fields are required.
type ReserveAppScreenshotRequest struct {
	ScreenshotSetID string // required — parent appScreenshotSets id
	FileName        string // required — e.g. "marketing-1.png"
	FileSize        int64  // required — total byte length of the source file
}

// Reserve is the first step of the upload flow. Apple responds with
// the created appScreenshots resource plus its uploadOperations array.
func (s *AppScreenshotsService) Reserve(ctx context.Context, req ReserveAppScreenshotRequest) (*AppScreenshot, error) {
	if req.ScreenshotSetID == "" || req.FileName == "" || req.FileSize <= 0 {
		return nil, &ClientError{Message: "AppScreenshots.Reserve: ScreenshotSetID, FileName, and positive FileSize are required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "appScreenshots",
			"attributes": map[string]any{
				"fileName": req.FileName,
				"fileSize": req.FileSize,
			},
			"relationships": map[string]any{
				"appScreenshotSet": map[string]any{
					"data": map[string]any{"type": "appScreenshotSets", "id": req.ScreenshotSetID},
				},
			},
		},
	}
	var doc Document[AppScreenshotAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/appScreenshots", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Get fetches a single screenshot by resource id.
func (s *AppScreenshotsService) Get(ctx context.Context, id string, query *Query) (*AppScreenshot, error) {
	if id == "" {
		return nil, &ClientError{Message: "AppScreenshots.Get: id is required"}
	}
	var doc Document[AppScreenshotAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/appScreenshots/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// ExecuteUploadOperations runs every [UploadOperation] from a reservation
// against the caller's source bytes. Apple streams to S3, so these
// requests use an unauthenticated [http.Client] — the auth is already
// embedded in each operation's URL + headers.
//
// The source slice must be at least long enough to cover the greatest
// offset+length pair Apple asked for.
func (s *AppScreenshotsService) ExecuteUploadOperations(ctx context.Context, ops []UploadOperation, source []byte) error {
	if len(ops) == 0 {
		return &ClientError{Message: "AppScreenshots.ExecuteUploadOperations: no upload operations"}
	}
	for i, op := range ops {
		if op.Offset < 0 || op.Length <= 0 || op.Offset+op.Length > int64(len(source)) {
			return &ClientError{Message: "AppScreenshots.ExecuteUploadOperations: upload operation out of bounds"}
		}
		slice := source[op.Offset : op.Offset+op.Length]
		method := op.Method
		if method == "" {
			method = "PUT"
		}
		req, err := http.NewRequestWithContext(ctx, method, op.URL, bytes.NewReader(slice))
		if err != nil {
			return &ClientError{Message: "AppScreenshots.ExecuteUploadOperations: build request", Cause: err}
		}
		for _, h := range op.Headers {
			req.Header.Set(h.Name, h.Value)
		}
		// Apple's docs state the upload target is authenticated via the
		// URL / Headers Apple gave us; do not attach App Store Connect
		// JWT here.
		resp, err := s.svc.httpClient.Do(req)
		if err != nil {
			return &ClientError{Message: "AppScreenshots.ExecuteUploadOperations: execute upload", Cause: err}
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return &ClientError{Message: "AppScreenshots.ExecuteUploadOperations: read upload response", Cause: readErr}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &ClientError{
				Message: "AppScreenshots.ExecuteUploadOperations: upload rejected",
				Cause: &APIError{
					StatusCode: resp.StatusCode,
					Errors: []Error{{
						Status: resp.Status,
						Title:  "upload operation " + strconv.Itoa(i) + " failed",
						Detail: string(body),
					}},
				},
			}
		}
	}
	return nil
}

// Commit is the final step of the upload flow. It marks the
// reservation as complete by setting uploaded=true and attaching the
// MD5 checksum of the full source file so Apple can verify integrity.
func (s *AppScreenshotsService) Commit(ctx context.Context, id string, checksum string) (*AppScreenshot, error) {
	if id == "" {
		return nil, &ClientError{Message: "AppScreenshots.Commit: id is required"}
	}
	if checksum == "" {
		return nil, &ClientError{Message: "AppScreenshots.Commit: checksum is required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "appScreenshots",
			"id":   id,
			"attributes": map[string]any{
				"uploaded":           true,
				"sourceFileChecksum": checksum,
			},
		},
	}
	var doc Document[AppScreenshotAttributes]
	if _, err := s.svc.do(ctx, "PATCH", "/v1/appScreenshots/"+id, nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete removes a screenshot by resource id.
func (s *AppScreenshotsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "AppScreenshots.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/appScreenshots/"+id, nil, nil, nil)
	return err
}

// Upload runs the full reserve → execute → commit flow for a single
// in-memory image. It is the convenience path for callers that already
// have the image bytes and a filename.
//
// Returns the committed screenshot resource. On any step failure Apple
// leaves the reservation in a "pending" state; callers may want to
// Delete the returned id to clean up.
func (s *AppScreenshotsService) Upload(ctx context.Context, screenshotSetID, fileName string, data []byte) (*AppScreenshot, error) {
	if screenshotSetID == "" || fileName == "" {
		return nil, &ClientError{Message: "AppScreenshots.Upload: screenshotSetID and fileName are required"}
	}
	if len(data) == 0 {
		return nil, &ClientError{Message: "AppScreenshots.Upload: data is empty"}
	}
	reservation, err := s.Reserve(ctx, ReserveAppScreenshotRequest{
		ScreenshotSetID: screenshotSetID,
		FileName:        fileName,
		FileSize:        int64(len(data)),
	})
	if err != nil {
		return nil, err
	}
	if err := s.ExecuteUploadOperations(ctx, reservation.Attributes.UploadOperations, data); err != nil {
		return reservation, err
	}
	sum := md5.Sum(data)
	return s.Commit(ctx, reservation.Id, hex.EncodeToString(sum[:]))
}

// itoa is a tiny strconv-free helper — used only when building error
// strings for upload failures, so pulling in strconv would be overkill.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
