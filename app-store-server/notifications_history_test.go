package AppStoreServer

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	Apple "github.com/godrealms/go-apple-sdk"
)

func testKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

// newServerTestClient returns an *Apple.Client whose App Store Server requests
// are redirected to an httptest server, so legacy endpoints can be tested
// without contacting Apple.
func newServerTestClient(t *testing.T, handler http.Handler) *Apple.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return Apple.NewClient(false, "KID", "ISS", "BID", testKeyPEM(t),
		Apple.WithServiceBaseURL(Apple.AppStoreServerClient, srv.URL))
}

// TestGetNotificationHistory_DecodesFieldsAndUsesQueryToken guards the empty-
// shell bug: the response must decode into real fields, the paginationToken
// must travel as a query parameter, and the filter criteria in the body.
func TestGetNotificationHistory_DecodesFieldsAndUsesQueryToken(t *testing.T) {
	var gotToken string
	var gotBody []byte
	c := newServerTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.URL.Query().Get("paginationToken")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"paginationToken":"NEXT",
			"hasMore":true,
			"notificationHistory":[
				{"signedPayload":"eyJ","sendAttempts":[{"attemptDate":1700000000123,"sendAttemptResult":"SUCCESS"}]}
			]
		}`))
	}))

	req := NotificationHistoryRequest{
		StartDate:        1600000000000,
		EndDate:          1600000100000,
		NotificationType: "SUBSCRIBED",
	}
	resp, err := GetNotificationHistory(context.Background(), c, req, "TOKEN")
	if err != nil {
		t.Fatalf("GetNotificationHistory: %v", err)
	}
	if gotToken != "TOKEN" {
		t.Errorf("paginationToken must be a query param, got query=%q", gotToken)
	}
	if !bytes.Contains(gotBody, []byte("startDate")) {
		t.Errorf("request body must carry filter fields, got %s", gotBody)
	}
	if resp.PaginationToken != "NEXT" {
		t.Errorf("PaginationToken not decoded: %q", resp.PaginationToken)
	}
	if !bool(resp.HasMore) {
		t.Error("HasMore not decoded")
	}
	if len(resp.NotificationHistory) != 1 {
		t.Fatalf("notificationHistory not decoded: %#v", resp.NotificationHistory)
	}
	if resp.NotificationHistory[0].SignedPayload != "eyJ" {
		t.Error("signedPayload not decoded")
	}
	if len(resp.NotificationHistory[0].SendAttempts) != 1 {
		t.Error("sendAttempts not decoded")
	}
}
