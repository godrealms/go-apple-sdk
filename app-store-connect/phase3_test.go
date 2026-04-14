package AppStoreConnect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// readBodyAsMap reads the request body as JSON into a generic map.
// Fails the test on decode error. Returns nil for empty bodies.
func readBodyAsMap(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return m
}

// ---- Builds ----

func TestBuildsService_List(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/builds" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "builds_list.json"))
	}))
	resp, err := svc.Builds().List(context.Background(), NewQuery().Limit(2))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(Data) = %d", len(resp.Data))
	}
	if resp.Data[0].Attributes.Version != "100" {
		t.Errorf("Version = %q", resp.Data[0].Attributes.Version)
	}
}

func TestBuildsService_Update(t *testing.T) {
	var captured struct {
		method string
		path   string
		body   map[string]any
	}
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.body = readBodyAsMap(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "build_updated.json"))
	}))
	build, err := svc.Builds().Update(context.Background(), "build-1", NewBuildUpdate().Expired(true))
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if captured.method != "PATCH" || captured.path != "/v1/builds/build-1" {
		t.Errorf("method=%q path=%q", captured.method, captured.path)
	}
	data, _ := captured.body["data"].(map[string]any)
	attrs, _ := data["attributes"].(map[string]any)
	if attrs["expired"] != true {
		t.Errorf("attributes.expired = %v", attrs["expired"])
	}
	if build.Id != "build-1" {
		t.Errorf("id = %q", build.Id)
	}
}

func TestBuildsService_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.Builds().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if _, err := svc.Builds().Update(context.Background(), "", NewBuildUpdate().Expired(true)); err == nil {
		t.Error("expected error for empty Update id")
	}
	if _, err := svc.Builds().Update(context.Background(), "b", nil); err == nil {
		t.Error("expected error for nil update")
	}
	if _, err := svc.Builds().Update(context.Background(), "b", NewBuildUpdate()); err == nil {
		t.Error("expected error for empty update")
	}
	// Other setter branches for coverage.
	u := NewBuildUpdate().UsesNonExemptEncryption(false).Set("custom", 1)
	if u.IsEmpty() {
		t.Error("IsEmpty = true after setters")
	}
}

// ---- BetaGroups ----

func TestBetaGroupsService_CreateAndRelationships(t *testing.T) {
	var seen []string
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/betaGroups":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			if data["type"] != "betaGroups" {
				t.Errorf("type = %v", data["type"])
			}
			rels, _ := data["relationships"].(map[string]any)
			app, _ := rels["app"].(map[string]any)
			appData, _ := app["data"].(map[string]any)
			if appData["id"] != "1234567890" {
				t.Errorf("app id = %v", appData["id"])
			}
			_, _ = w.Write(loadFixture(t, "beta_group_created.json"))
		case "/v1/betaGroups/bg-1/relationships/builds":
			body := readBodyAsMap(t, r)
			arr, _ := body["data"].([]any)
			if len(arr) != 2 {
				t.Errorf("builds linkage len = %d", len(arr))
			}
			w.WriteHeader(http.StatusNoContent)
		case "/v1/betaGroups/bg-1/relationships/betaTesters":
			w.WriteHeader(http.StatusNoContent)
		case "/v1/betaGroups/bg-1":
			if r.Method != "DELETE" {
				t.Errorf("method = %q", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	trueVal := true
	bg, err := svc.BetaGroups().Create(context.Background(), CreateBetaGroupRequest{
		AppID: "1234567890", Name: "QA Team",
		PublicLinkEnabled: &trueVal,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if bg.Id != "bg-1" {
		t.Errorf("id = %q", bg.Id)
	}
	if err := svc.BetaGroups().AddBuilds(context.Background(), "bg-1", []string{"build-1", "build-2"}); err != nil {
		t.Fatalf("AddBuilds: %v", err)
	}
	if err := svc.BetaGroups().AddBetaTesters(context.Background(), "bg-1", []string{"tester-1"}); err != nil {
		t.Fatalf("AddBetaTesters: %v", err)
	}
	if err := svc.BetaGroups().Delete(context.Background(), "bg-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(seen) != 4 {
		t.Errorf("seen = %v", seen)
	}
}

func TestBetaGroupsService_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.BetaGroups().Create(context.Background(), CreateBetaGroupRequest{Name: "x"}); err == nil {
		t.Error("expected error for missing AppID")
	}
	if _, err := svc.BetaGroups().Create(context.Background(), CreateBetaGroupRequest{AppID: "a"}); err == nil {
		t.Error("expected error for missing Name")
	}
	if _, err := svc.BetaGroups().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if err := svc.BetaGroups().AddBuilds(context.Background(), "", []string{"b"}); err == nil {
		t.Error("expected error for empty group id in AddBuilds")
	}
	if err := svc.BetaGroups().AddBuilds(context.Background(), "g", nil); err == nil {
		t.Error("expected error for empty build ids")
	}
	if err := svc.BetaGroups().AddBetaTesters(context.Background(), "", []string{"t"}); err == nil {
		t.Error("expected error for empty group id in AddBetaTesters")
	}
	if err := svc.BetaGroups().AddBetaTesters(context.Background(), "g", nil); err == nil {
		t.Error("expected error for empty tester ids")
	}
	if err := svc.BetaGroups().Delete(context.Background(), ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- BetaTesterInvitations ----

func TestBetaTesterInvitations_Create(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/betaTesterInvitations" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		body := readBodyAsMap(t, r)
		data, _ := body["data"].(map[string]any)
		rels, _ := data["relationships"].(map[string]any)
		app, _ := rels["app"].(map[string]any)
		appData, _ := app["data"].(map[string]any)
		if appData["id"] != "app-1" {
			t.Errorf("app id = %v", appData["id"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"type":"betaTesterInvitations","id":"inv-1"}}`))
	}))
	inv, err := svc.BetaTesterInvitations().Create(context.Background(), "app-1", "tester-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if inv.Id != "inv-1" {
		t.Errorf("id = %q", inv.Id)
	}
	if _, err := svc.BetaTesterInvitations().Create(context.Background(), "", "t"); err == nil {
		t.Error("expected error for empty appID")
	}
	if _, err := svc.BetaTesterInvitations().Create(context.Background(), "a", ""); err == nil {
		t.Error("expected error for empty testerID")
	}
}

// ---- BundleIDs ----

func TestBundleIDsService_Create(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/bundleIds" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body := readBodyAsMap(t, r)
		data, _ := body["data"].(map[string]any)
		attrs, _ := data["attributes"].(map[string]any)
		if attrs["identifier"] != "com.acme.widgets" || attrs["platform"] != "IOS" {
			t.Errorf("attrs = %+v", attrs)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "bundle_id_created.json"))
	}))
	bid, err := svc.BundleIDs().Create(context.Background(), CreateBundleIDRequest{
		Identifier: "com.acme.widgets",
		Name:       "Acme Widgets",
		Platform:   "IOS",
		SeedID:     "ABC123XYZ0",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if bid.Attributes.Identifier != "com.acme.widgets" {
		t.Errorf("Identifier = %q", bid.Attributes.Identifier)
	}
}

func TestBundleIDsService_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.BundleIDs().Create(context.Background(), CreateBundleIDRequest{}); err == nil {
		t.Error("expected error for empty request")
	}
	if _, err := svc.BundleIDs().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if err := svc.BundleIDs().Delete(context.Background(), ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- Certificates ----

func TestCertificatesService_Create(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/certificates" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body := readBodyAsMap(t, r)
		data, _ := body["data"].(map[string]any)
		attrs, _ := data["attributes"].(map[string]any)
		if attrs["csrContent"] != "CSR_B64" || attrs["certificateType"] != "IOS_DISTRIBUTION" {
			t.Errorf("attrs = %+v", attrs)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "certificate_created.json"))
	}))
	cert, err := svc.Certificates().Create(context.Background(), CreateCertificateRequest{
		CSRContent:      "CSR_B64",
		CertificateType: "IOS_DISTRIBUTION",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if cert.Attributes.SerialNumber != "01A2B3C4D5" {
		t.Errorf("SerialNumber = %q", cert.Attributes.SerialNumber)
	}
}

func TestCertificatesService_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.Certificates().Create(context.Background(), CreateCertificateRequest{}); err == nil {
		t.Error("expected error for empty request")
	}
	if _, err := svc.Certificates().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if err := svc.Certificates().Delete(context.Background(), ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- Profiles ----

func TestProfilesService_Create(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/profiles" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body := readBodyAsMap(t, r)
		data, _ := body["data"].(map[string]any)
		rels, _ := data["relationships"].(map[string]any)
		if _, ok := rels["bundleId"]; !ok {
			t.Error("missing bundleId relationship")
		}
		certs, _ := rels["certificates"].(map[string]any)
		certArr, _ := certs["data"].([]any)
		if len(certArr) != 1 {
			t.Errorf("cert linkage len = %d", len(certArr))
		}
		devs, _ := rels["devices"].(map[string]any)
		devArr, _ := devs["data"].([]any)
		if len(devArr) != 2 {
			t.Errorf("device linkage len = %d", len(devArr))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "profile_created.json"))
	}))
	p, err := svc.Profiles().Create(context.Background(), CreateProfileRequest{
		Name:           "Acme Distribution",
		ProfileType:    "IOS_APP_STORE",
		BundleID:       "bid-1",
		CertificateIDs: []string{"cert-1"},
		DeviceIDs:      []string{"dev-1", "dev-2"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Attributes.UUID == "" {
		t.Error("expected UUID to decode")
	}
}

func TestProfilesService_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.Profiles().Create(context.Background(), CreateProfileRequest{Name: "x"}); err == nil {
		t.Error("expected error for missing required fields")
	}
	if _, err := svc.Profiles().Create(context.Background(), CreateProfileRequest{
		Name: "x", ProfileType: "y", BundleID: "z",
	}); err == nil {
		t.Error("expected error for missing certs")
	}
	if _, err := svc.Profiles().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if err := svc.Profiles().Delete(context.Background(), ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- Users ----

func TestUsersService_ListAndUpdate(t *testing.T) {
	var seen []string
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/users":
			_, _ = w.Write(loadFixture(t, "users_list.json"))
		case "/v1/users/user-2":
			if r.Method == "PATCH" {
				body := readBodyAsMap(t, r)
				data, _ := body["data"].(map[string]any)
				attrs, _ := data["attributes"].(map[string]any)
				roles, _ := attrs["roles"].([]any)
				if len(roles) != 2 {
					t.Errorf("roles len = %d", len(roles))
				}
				_, _ = w.Write([]byte(`{"data":{"type":"users","id":"user-2","attributes":{"roles":["DEVELOPER","APP_MANAGER"]}}}`))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	list, err := svc.Users().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 2 {
		t.Fatalf("len = %d", len(list.Data))
	}

	upd := NewUserUpdate().
		Roles("DEVELOPER", "APP_MANAGER").
		AllAppsVisible(true).
		ProvisioningAllowed(false)
	u, err := svc.Users().Update(context.Background(), "user-2", upd)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(u.Attributes.Roles) != 2 {
		t.Errorf("roles = %v", u.Attributes.Roles)
	}
	if err := svc.Users().Delete(context.Background(), "user-2"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(seen) != 3 {
		t.Errorf("seen = %v", seen)
	}
}

func TestUsersService_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.Users().Get(context.Background(), "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if _, err := svc.Users().Update(context.Background(), "", NewUserUpdate().AllAppsVisible(true)); err == nil {
		t.Error("expected error for empty Update id")
	}
	if _, err := svc.Users().Update(context.Background(), "u", NewUserUpdate()); err == nil {
		t.Error("expected error for empty update")
	}
	if err := svc.Users().Delete(context.Background(), ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- UserInvitations ----

func TestUserInvitationsService_Create(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/userInvitations" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body := readBodyAsMap(t, r)
		data, _ := body["data"].(map[string]any)
		attrs, _ := data["attributes"].(map[string]any)
		if attrs["email"] != "carol@example.com" {
			t.Errorf("email = %v", attrs["email"])
		}
		rels, _ := data["relationships"].(map[string]any)
		if rels == nil {
			t.Error("expected visibleApps relationship")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "user_invitation_created.json"))
	}))
	fa := false
	inv, err := svc.UserInvitations().Create(context.Background(), CreateUserInvitationRequest{
		Email:               "carol@example.com",
		FirstName:           "Carol",
		LastName:            "Cherry",
		Roles:               []string{"DEVELOPER"},
		AllAppsVisible:      &fa,
		ProvisioningAllowed: &fa,
		VisibleAppIDs:       []string{"app-1"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if inv.Id != "inv-1" {
		t.Errorf("id = %q", inv.Id)
	}
}

func TestUserInvitationsService_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	if _, err := svc.UserInvitations().Create(context.Background(), CreateUserInvitationRequest{}); err == nil {
		t.Error("expected error for empty request")
	}
	if _, err := svc.UserInvitations().Create(context.Background(), CreateUserInvitationRequest{
		Email: "a@b", FirstName: "A", LastName: "B",
	}); err == nil {
		t.Error("expected error for missing roles")
	}
	if err := svc.UserInvitations().Delete(context.Background(), ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- Helpers ----

func TestBuildIdentifiersHelper(t *testing.T) {
	out := buildIdentifiers("builds", []string{"a", "b"})
	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0]["type"] != "builds" || out[0]["id"] != "a" {
		t.Errorf("out[0] = %+v", out[0])
	}
}

// A list/iterate smoke test to exercise paginator code paths across
// the new services. Uses users_list.json as a terminal page.
func TestUsersService_ListIteratorSmoke(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/users") {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "users_list.json"))
	}))
	it := svc.Users().ListIterator(nil)
	pages := 0
	for it.Next(context.Background()) {
		pages++
		if pages > 1 {
			break
		}
	}
	if err := it.Err(); err != nil {
		t.Fatalf("err: %v", err)
	}
}
