package AppStoreConnect

import (
	"context"
	"time"
)

// CertificatesService provides access to /v1/certificates, the signing
// certificate catalog used for code-signing and push notifications.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/certificates
type CertificatesService struct {
	svc *Service
}

// Certificate is a typed alias for a JSON:API certificates resource.
type Certificate = Resource[CertificateAttributes]

// CertificateAttributes models the attributes of a certificates resource.
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/certificate/attributes
type CertificateAttributes struct {
	Name                string     `json:"name,omitempty"`
	CertificateType     string     `json:"certificateType,omitempty"`
	DisplayName         string     `json:"displayName,omitempty"`
	SerialNumber        string     `json:"serialNumber,omitempty"`
	Platform            string     `json:"platform,omitempty"`
	ExpirationDate      *time.Time `json:"expirationDate,omitempty"`
	CertificateContent  string     `json:"certificateContent,omitempty"` // base64 DER
	CsrContent          string     `json:"csrContent,omitempty"`
}

// ListCertificatesResponse is the decoded response for [CertificatesService.List].
type ListCertificatesResponse struct {
	Data     []Certificate   `json:"data"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
}

// List returns a page of certificates matching the query.
func (s *CertificatesService) List(ctx context.Context, query *Query) (*ListCertificatesResponse, error) {
	var doc Document[CertificateAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/certificates", query, nil, &doc); err != nil {
		return nil, err
	}
	data, err := doc.AsCollection()
	if err != nil {
		return nil, err
	}
	return &ListCertificatesResponse{Data: data, Included: doc.Included, Links: doc.Links}, nil
}

// ListIterator returns a paginator that walks every page of certificates.
func (s *CertificatesService) ListIterator(query *Query) *Paginator[CertificateAttributes] {
	return newPaginator[CertificateAttributes](s.svc, "/v1/certificates", query)
}

// Get fetches a single certificate by resource id.
func (s *CertificatesService) Get(ctx context.Context, id string, query *Query) (*Certificate, error) {
	if id == "" {
		return nil, &ClientError{Message: "Certificates.Get: id is required"}
	}
	var doc Document[CertificateAttributes]
	if _, err := s.svc.do(ctx, "GET", "/v1/certificates/"+id, query, nil, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// CreateCertificateRequest describes a new certificate to sign.
// CSRContent must be the base64-encoded contents of a CSR file
// produced via openssl / Keychain Access; CertificateType must be
// one of the values in Apple's catalog (e.g. IOS_DISTRIBUTION,
// IOS_DEVELOPMENT, DEVELOPER_ID_APPLICATION).
type CreateCertificateRequest struct {
	CSRContent      string // required, base64
	CertificateType string // required, see Apple's enum
}

// Create submits a CSR and returns the issued certificate.
// See https://developer.apple.com/documentation/appstoreconnectapi/create_a_certificate
func (s *CertificatesService) Create(ctx context.Context, req CreateCertificateRequest) (*Certificate, error) {
	if req.CSRContent == "" || req.CertificateType == "" {
		return nil, &ClientError{Message: "Certificates.Create: CSRContent and CertificateType are required"}
	}
	body := map[string]any{
		"data": map[string]any{
			"type": "certificates",
			"attributes": map[string]any{
				"csrContent":      req.CSRContent,
				"certificateType": req.CertificateType,
			},
		},
	}
	var doc Document[CertificateAttributes]
	if _, err := s.svc.do(ctx, "POST", "/v1/certificates", nil, body, &doc); err != nil {
		return nil, err
	}
	return doc.AsResource()
}

// Delete revokes a certificate. Apple returns 204 on success; once
// revoked the certificate's signatures continue to be valid but it
// can no longer be used for new signing operations.
func (s *CertificatesService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ClientError{Message: "Certificates.Delete: id is required"}
	}
	_, err := s.svc.do(ctx, "DELETE", "/v1/certificates/"+id, nil, nil, nil)
	return err
}
