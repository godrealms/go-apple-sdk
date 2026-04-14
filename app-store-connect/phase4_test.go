package AppStoreConnect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---- AppStoreVersions ----

func TestAppStoreVersions_CreateAndLifecycle(t *testing.T) {
	var seen []string
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/appStoreVersions":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["versionString"] != "1.2.3" {
				t.Errorf("versionString = %v", attrs["versionString"])
			}
			rels, _ := data["relationships"].(map[string]any)
			app, _ := rels["app"].(map[string]any)
			appData, _ := app["data"].(map[string]any)
			if appData["id"] != "1234567890" {
				t.Errorf("app id = %v", appData["id"])
			}
			_, _ = w.Write(loadFixture(t, "app_store_version_created.json"))
		case r.Method == "GET" && r.URL.Path == "/v1/apps/1234567890/appStoreVersions":
			_, _ = w.Write(loadFixture(t, "app_store_versions_list.json"))
		case r.Method == "GET" && r.URL.Path == "/v1/appStoreVersions/ver-1":
			_, _ = w.Write(loadFixture(t, "app_store_version_created.json"))
		case r.Method == "PATCH" && r.URL.Path == "/v1/appStoreVersions/ver-1":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["copyright"] != "© 2026 Acme" {
				t.Errorf("copyright = %v", attrs["copyright"])
			}
			_, _ = w.Write(loadFixture(t, "app_store_version_created.json"))
		case r.Method == "PATCH" && r.URL.Path == "/v1/appStoreVersions/ver-1/relationships/build":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			if data["type"] != "builds" || data["id"] != "build-42" {
				t.Errorf("build linkage = %+v", data)
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "DELETE" && r.URL.Path == "/v1/appStoreVersions/ver-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()

	ver, err := svc.AppStoreVersions().Create(ctx, CreateAppStoreVersionRequest{
		AppID:         "1234567890",
		Platform:      "IOS",
		VersionString: "1.2.3",
		Copyright:     "© 2026 Acme",
		ReleaseType:   "MANUAL",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ver.Id != "ver-1" || ver.Attributes.VersionString != "1.2.3" {
		t.Errorf("ver = %+v", ver)
	}

	list, err := svc.AppStoreVersions().ListForApp(ctx, "1234567890", NewQuery().Limit(10))
	if err != nil {
		t.Fatalf("ListForApp: %v", err)
	}
	if len(list.Data) != 2 {
		t.Errorf("list len = %d", len(list.Data))
	}

	if _, err := svc.AppStoreVersions().Get(ctx, "ver-1", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}

	upd := NewAppStoreVersionUpdate().
		VersionString("1.2.3").
		Copyright("© 2026 Acme").
		ReleaseType("AFTER_APPROVAL").
		EarliestReleaseDate(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)).
		Downloadable(true).
		Set("custom", "v")
	if upd.IsEmpty() {
		t.Error("expected non-empty update")
	}
	if _, err := svc.AppStoreVersions().Update(ctx, "ver-1", upd); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := svc.AppStoreVersions().SelectBuild(ctx, "ver-1", "build-42"); err != nil {
		t.Fatalf("SelectBuild: %v", err)
	}

	if err := svc.AppStoreVersions().Delete(ctx, "ver-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if len(seen) != 6 {
		t.Errorf("seen = %v", seen)
	}
}

func TestAppStoreVersions_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	ctx := context.Background()

	if _, err := svc.AppStoreVersions().Create(ctx, CreateAppStoreVersionRequest{}); err == nil {
		t.Error("expected error for empty Create")
	}
	if _, err := svc.AppStoreVersions().ListForApp(ctx, "", nil); err == nil {
		t.Error("expected error for empty appID in ListForApp")
	}
	if _, err := svc.AppStoreVersions().Get(ctx, "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if _, err := svc.AppStoreVersions().Update(ctx, "", NewAppStoreVersionUpdate().Copyright("c")); err == nil {
		t.Error("expected error for empty Update id")
	}
	if _, err := svc.AppStoreVersions().Update(ctx, "v", nil); err == nil {
		t.Error("expected error for nil update")
	}
	if _, err := svc.AppStoreVersions().Update(ctx, "v", NewAppStoreVersionUpdate()); err == nil {
		t.Error("expected error for empty update")
	}
	if err := svc.AppStoreVersions().Delete(ctx, ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
	if err := svc.AppStoreVersions().SelectBuild(ctx, "", "b"); err == nil {
		t.Error("expected error for empty SelectBuild versionID")
	}
	if err := svc.AppStoreVersions().SelectBuild(ctx, "v", ""); err == nil {
		t.Error("expected error for empty SelectBuild buildID")
	}
	// Iterator non-nil.
	if svc.AppStoreVersions().ListForAppIterator("app-1", nil) == nil {
		t.Error("ListForAppIterator nil")
	}
}

// ---- AppStoreVersionSubmissions ----

func TestAppStoreVersionSubmissions(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/appStoreVersionSubmissions":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			rels, _ := data["relationships"].(map[string]any)
			ver, _ := rels["appStoreVersion"].(map[string]any)
			verData, _ := ver["data"].(map[string]any)
			if verData["id"] != "ver-1" {
				t.Errorf("version id = %v", verData["id"])
			}
			_, _ = w.Write(loadFixture(t, "app_store_version_submission_created.json"))
		case r.Method == "DELETE" && r.URL.Path == "/v1/appStoreVersionSubmissions/sub-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()
	sub, err := svc.AppStoreVersionSubmissions().Submit(ctx, "ver-1")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if sub.Id != "sub-1" {
		t.Errorf("sub id = %q", sub.Id)
	}
	if err := svc.AppStoreVersionSubmissions().Cancel(ctx, "sub-1"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	if _, err := svc.AppStoreVersionSubmissions().Submit(ctx, ""); err == nil {
		t.Error("expected error for empty versionID")
	}
	if err := svc.AppStoreVersionSubmissions().Cancel(ctx, ""); err == nil {
		t.Error("expected error for empty submissionID")
	}
}

// ---- AppStoreVersionLocalizations ----

func TestAppStoreVersionLocalizations(t *testing.T) {
	var seen []string
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/appStoreVersionLocalizations":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["locale"] != "en-US" {
				t.Errorf("locale = %v", attrs["locale"])
			}
			if attrs["whatsNew"] == nil {
				t.Error("expected whatsNew in body")
			}
			rels, _ := data["relationships"].(map[string]any)
			ver, _ := rels["appStoreVersion"].(map[string]any)
			verData, _ := ver["data"].(map[string]any)
			if verData["id"] != "ver-1" {
				t.Errorf("version id = %v", verData["id"])
			}
			_, _ = w.Write(loadFixture(t, "app_store_version_localization_created.json"))
		case r.Method == "GET" && r.URL.Path == "/v1/appStoreVersions/ver-1/appStoreVersionLocalizations":
			_, _ = w.Write([]byte(`{"data":[{"type":"appStoreVersionLocalizations","id":"loc-1","attributes":{"locale":"en-US"}}]}`))
		case r.Method == "GET" && r.URL.Path == "/v1/appStoreVersionLocalizations/loc-1":
			_, _ = w.Write(loadFixture(t, "app_store_version_localization_created.json"))
		case r.Method == "PATCH" && r.URL.Path == "/v1/appStoreVersionLocalizations/loc-1":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["promotionalText"] != "Hot new thing" {
				t.Errorf("promotionalText = %v", attrs["promotionalText"])
			}
			_, _ = w.Write(loadFixture(t, "app_store_version_localization_created.json"))
		case r.Method == "DELETE" && r.URL.Path == "/v1/appStoreVersionLocalizations/loc-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()
	loc, err := svc.AppStoreVersionLocalizations().Create(ctx, CreateAppStoreVersionLocalizationRequest{
		VersionID:       "ver-1",
		Locale:          "en-US",
		Description:     "An app that does things",
		Keywords:        "things,stuff",
		MarketingURL:    "https://example.com/marketing",
		PromotionalText: "Fresh!",
		SupportURL:      "https://example.com/support",
		WhatsNew:        "- Everything is new",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if loc.Attributes.Locale != "en-US" {
		t.Errorf("locale = %q", loc.Attributes.Locale)
	}

	list, err := svc.AppStoreVersionLocalizations().ListForVersion(ctx, "ver-1", nil)
	if err != nil {
		t.Fatalf("ListForVersion: %v", err)
	}
	if len(list.Data) != 1 {
		t.Errorf("list len = %d", len(list.Data))
	}

	if _, err := svc.AppStoreVersionLocalizations().Get(ctx, "loc-1", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}

	upd := NewAppStoreVersionLocalizationUpdate().
		Description("new desc").
		Keywords("new,kw").
		MarketingURL("https://example.com/new").
		PromotionalText("Hot new thing").
		SupportURL("https://example.com/support2").
		WhatsNew("new release").
		Set("custom", 1)
	if upd.IsEmpty() {
		t.Error("expected non-empty update")
	}
	if _, err := svc.AppStoreVersionLocalizations().Update(ctx, "loc-1", upd); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := svc.AppStoreVersionLocalizations().Delete(ctx, "loc-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(seen) != 5 {
		t.Errorf("seen = %v", seen)
	}
}

func TestAppStoreVersionLocalizations_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	ctx := context.Background()
	if _, err := svc.AppStoreVersionLocalizations().Create(ctx, CreateAppStoreVersionLocalizationRequest{}); err == nil {
		t.Error("expected error for empty Create")
	}
	if _, err := svc.AppStoreVersionLocalizations().ListForVersion(ctx, "", nil); err == nil {
		t.Error("expected error for empty versionID")
	}
	if _, err := svc.AppStoreVersionLocalizations().Get(ctx, "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if _, err := svc.AppStoreVersionLocalizations().Update(ctx, "", NewAppStoreVersionLocalizationUpdate().Keywords("k")); err == nil {
		t.Error("expected error for empty Update id")
	}
	if _, err := svc.AppStoreVersionLocalizations().Update(ctx, "l", nil); err == nil {
		t.Error("expected error for nil update")
	}
	if _, err := svc.AppStoreVersionLocalizations().Update(ctx, "l", NewAppStoreVersionLocalizationUpdate()); err == nil {
		t.Error("expected error for empty update")
	}
	if err := svc.AppStoreVersionLocalizations().Delete(ctx, ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- AppScreenshotSets ----

func TestAppScreenshotSets(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/appScreenshotSets":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["screenshotDisplayType"] != "APP_IPHONE_67" {
				t.Errorf("type = %v", attrs["screenshotDisplayType"])
			}
			_, _ = w.Write(loadFixture(t, "app_screenshot_set_created.json"))
		case r.Method == "GET" && r.URL.Path == "/v1/appStoreVersionLocalizations/loc-1/appScreenshotSets":
			_, _ = w.Write([]byte(`{"data":[{"type":"appScreenshotSets","id":"set-1","attributes":{"screenshotDisplayType":"APP_IPHONE_67"}}]}`))
		case r.Method == "GET" && r.URL.Path == "/v1/appScreenshotSets/set-1":
			_, _ = w.Write(loadFixture(t, "app_screenshot_set_created.json"))
		case r.Method == "DELETE" && r.URL.Path == "/v1/appScreenshotSets/set-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()
	set, err := svc.AppScreenshotSets().Create(ctx, CreateAppScreenshotSetRequest{
		LocalizationID:        "loc-1",
		ScreenshotDisplayType: "APP_IPHONE_67",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if set.Id != "set-1" {
		t.Errorf("id = %q", set.Id)
	}
	if _, err := svc.AppScreenshotSets().ListForLocalization(ctx, "loc-1", nil); err != nil {
		t.Fatalf("ListForLocalization: %v", err)
	}
	if _, err := svc.AppScreenshotSets().Get(ctx, "set-1", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if err := svc.AppScreenshotSets().Delete(ctx, "set-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Validation
	if _, err := svc.AppScreenshotSets().Create(ctx, CreateAppScreenshotSetRequest{}); err == nil {
		t.Error("expected error for empty Create")
	}
	if _, err := svc.AppScreenshotSets().ListForLocalization(ctx, "", nil); err == nil {
		t.Error("expected error for empty localizationID")
	}
	if _, err := svc.AppScreenshotSets().Get(ctx, "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if err := svc.AppScreenshotSets().Delete(ctx, ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- AppScreenshots (upload flow) ----

func TestAppScreenshots_UploadFlow(t *testing.T) {
	// Upload target (S3 stand-in) — Apple sends plain PUT with no auth.
	var uploadedBytes []byte
	uploadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("upload method = %q", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "image/png" {
			t.Errorf("upload content-type = %q", ct)
		}
		buf := make([]byte, 64)
		n, _ := r.Body.Read(buf)
		uploadedBytes = append(uploadedBytes, buf[:n]...)
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadSrv.Close()

	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/appScreenshots":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["fileName"] != "hero.png" {
				t.Errorf("fileName = %v", attrs["fileName"])
			}
			if _, ok := attrs["fileSize"]; !ok {
				t.Error("missing fileSize")
			}
			// Rewrite fixture URL placeholder with the test upload srv URL.
			fx := loadFixture(t, "app_screenshot_reserved.json")
			out := strings.ReplaceAll(string(fx), "{UPLOAD_URL}", uploadSrv.URL+"/upload")
			_, _ = w.Write([]byte(out))
		case r.Method == "PATCH" && r.URL.Path == "/v1/appScreenshots/shot-1":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["uploaded"] != true {
				t.Errorf("uploaded = %v", attrs["uploaded"])
			}
			if attrs["sourceFileChecksum"] == "" {
				t.Error("missing checksum")
			}
			_, _ = w.Write(loadFixture(t, "app_screenshot_committed.json"))
		case r.Method == "GET" && r.URL.Path == "/v1/appScreenshots/shot-1":
			_, _ = w.Write(loadFixture(t, "app_screenshot_committed.json"))
		case r.Method == "DELETE" && r.URL.Path == "/v1/appScreenshots/shot-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()
	// Payload length must match fileSize:12 in the fixture.
	png := []byte("hello,shot!!") // 12 bytes
	shot, err := svc.AppScreenshots().Upload(ctx, "set-1", "hero.png", png)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if shot.Id != "shot-1" {
		t.Errorf("id = %q", shot.Id)
	}
	if string(uploadedBytes) != string(png) {
		t.Errorf("upload payload = %q", uploadedBytes)
	}
	if shot.Attributes.Uploaded == nil || !*shot.Attributes.Uploaded {
		t.Error("expected uploaded=true")
	}

	// Drive Get and Delete for coverage.
	if _, err := svc.AppScreenshots().Get(ctx, "shot-1", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if err := svc.AppScreenshots().Delete(ctx, "shot-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestAppScreenshots_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	ctx := context.Background()

	if _, err := svc.AppScreenshots().Reserve(ctx, ReserveAppScreenshotRequest{}); err == nil {
		t.Error("expected error for empty Reserve")
	}
	if _, err := svc.AppScreenshots().Get(ctx, "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if _, err := svc.AppScreenshots().Commit(ctx, "", "chk"); err == nil {
		t.Error("expected error for empty Commit id")
	}
	if _, err := svc.AppScreenshots().Commit(ctx, "s", ""); err == nil {
		t.Error("expected error for empty Commit checksum")
	}
	if err := svc.AppScreenshots().Delete(ctx, ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
	if _, err := svc.AppScreenshots().Upload(ctx, "", "f", []byte("x")); err == nil {
		t.Error("expected error for empty setID in Upload")
	}
	if _, err := svc.AppScreenshots().Upload(ctx, "s", "f", nil); err == nil {
		t.Error("expected error for empty data in Upload")
	}
	// ExecuteUploadOperations validation.
	if err := svc.AppScreenshots().ExecuteUploadOperations(ctx, nil, []byte("x")); err == nil {
		t.Error("expected error for empty ops")
	}
	ops := []UploadOperation{{Method: "PUT", URL: "http://example", Offset: 0, Length: 10}}
	if err := svc.AppScreenshots().ExecuteUploadOperations(ctx, ops, []byte("short")); err == nil {
		t.Error("expected error for out-of-bounds op")
	}
}

// ---- InAppPurchases ----

func TestInAppPurchases(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/inAppPurchases":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["productId"] != "com.acme.widgets.pro" {
				t.Errorf("productId = %v", attrs["productId"])
			}
			if attrs["familySharable"] != true {
				t.Errorf("familySharable = %v", attrs["familySharable"])
			}
			_, _ = w.Write(loadFixture(t, "in_app_purchase_created.json"))
		case r.Method == "GET" && r.URL.Path == "/v1/apps/1234567890/inAppPurchasesV2":
			_, _ = w.Write([]byte(`{"data":[{"type":"inAppPurchases","id":"iap-1","attributes":{"productId":"com.acme.widgets.pro"}}]}`))
		case r.Method == "GET" && r.URL.Path == "/v1/inAppPurchases/iap-1":
			_, _ = w.Write(loadFixture(t, "in_app_purchase_created.json"))
		case r.Method == "PATCH" && r.URL.Path == "/v1/inAppPurchases/iap-1":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["name"] != "Pro Pack V2" {
				t.Errorf("name = %v", attrs["name"])
			}
			_, _ = w.Write(loadFixture(t, "in_app_purchase_created.json"))
		case r.Method == "DELETE" && r.URL.Path == "/v1/inAppPurchases/iap-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()
	tr := true
	iap, err := svc.InAppPurchases().Create(ctx, CreateInAppPurchaseRequest{
		AppID:             "1234567890",
		Name:              "Pro Pack",
		ProductID:         "com.acme.widgets.pro",
		InAppPurchaseType: "NON_CONSUMABLE",
		ReviewNote:        "please approve",
		FamilySharable:    &tr,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if iap.Id != "iap-1" {
		t.Errorf("id = %q", iap.Id)
	}
	if _, err := svc.InAppPurchases().ListForApp(ctx, "1234567890", nil); err != nil {
		t.Fatalf("ListForApp: %v", err)
	}
	if _, err := svc.InAppPurchases().Get(ctx, "iap-1", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}

	upd := NewInAppPurchaseUpdate().
		Name("Pro Pack V2").
		ReviewNote("updated").
		FamilySharable(false).
		Set("custom", 2)
	if upd.IsEmpty() {
		t.Error("expected non-empty update")
	}
	if _, err := svc.InAppPurchases().Update(ctx, "iap-1", upd); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := svc.InAppPurchases().Delete(ctx, "iap-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if svc.InAppPurchases().ListForAppIterator("1234567890", nil) == nil {
		t.Error("iterator nil")
	}
}

func TestInAppPurchases_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	ctx := context.Background()
	if _, err := svc.InAppPurchases().Create(ctx, CreateInAppPurchaseRequest{}); err == nil {
		t.Error("expected error for empty Create")
	}
	if _, err := svc.InAppPurchases().ListForApp(ctx, "", nil); err == nil {
		t.Error("expected error for empty appID")
	}
	if _, err := svc.InAppPurchases().Get(ctx, "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if _, err := svc.InAppPurchases().Update(ctx, "", NewInAppPurchaseUpdate().Name("x")); err == nil {
		t.Error("expected error for empty Update id")
	}
	if _, err := svc.InAppPurchases().Update(ctx, "i", nil); err == nil {
		t.Error("expected error for nil update")
	}
	if _, err := svc.InAppPurchases().Update(ctx, "i", NewInAppPurchaseUpdate()); err == nil {
		t.Error("expected error for empty update")
	}
	if err := svc.InAppPurchases().Delete(ctx, ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}

// ---- SubscriptionGroups ----

func TestSubscriptionGroups(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/subscriptionGroups":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["referenceName"] != "Pro Plans" {
				t.Errorf("referenceName = %v", attrs["referenceName"])
			}
			_, _ = w.Write(loadFixture(t, "subscription_group_created.json"))
		case r.Method == "GET" && r.URL.Path == "/v1/apps/1234567890/subscriptionGroups":
			_, _ = w.Write([]byte(`{"data":[{"type":"subscriptionGroups","id":"grp-1","attributes":{"referenceName":"Pro Plans"}}]}`))
		case r.Method == "GET" && r.URL.Path == "/v1/subscriptionGroups/grp-1":
			_, _ = w.Write(loadFixture(t, "subscription_group_created.json"))
		case r.Method == "PATCH" && r.URL.Path == "/v1/subscriptionGroups/grp-1":
			body := readBodyAsMap(t, r)
			data, _ := body["data"].(map[string]any)
			attrs, _ := data["attributes"].(map[string]any)
			if attrs["referenceName"] != "Pro Plans Renamed" {
				t.Errorf("referenceName = %v", attrs["referenceName"])
			}
			_, _ = w.Write(loadFixture(t, "subscription_group_created.json"))
		case r.Method == "DELETE" && r.URL.Path == "/v1/subscriptionGroups/grp-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()
	grp, err := svc.SubscriptionGroups().Create(ctx, CreateSubscriptionGroupRequest{
		AppID:         "1234567890",
		ReferenceName: "Pro Plans",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if grp.Id != "grp-1" {
		t.Errorf("id = %q", grp.Id)
	}
	if _, err := svc.SubscriptionGroups().ListForApp(ctx, "1234567890", nil); err != nil {
		t.Fatalf("ListForApp: %v", err)
	}
	if _, err := svc.SubscriptionGroups().Get(ctx, "grp-1", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, err := svc.SubscriptionGroups().Update(ctx, "grp-1", "Pro Plans Renamed"); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := svc.SubscriptionGroups().Delete(ctx, "grp-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if svc.SubscriptionGroups().ListForAppIterator("1234567890", nil) == nil {
		t.Error("iterator nil")
	}
}

func TestSubscriptionGroups_Validation(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	ctx := context.Background()
	if _, err := svc.SubscriptionGroups().Create(ctx, CreateSubscriptionGroupRequest{}); err == nil {
		t.Error("expected error for empty Create")
	}
	if _, err := svc.SubscriptionGroups().ListForApp(ctx, "", nil); err == nil {
		t.Error("expected error for empty appID")
	}
	if _, err := svc.SubscriptionGroups().Get(ctx, "", nil); err == nil {
		t.Error("expected error for empty Get id")
	}
	if _, err := svc.SubscriptionGroups().Update(ctx, "", "x"); err == nil {
		t.Error("expected error for empty Update id")
	}
	if _, err := svc.SubscriptionGroups().Update(ctx, "g", ""); err == nil {
		t.Error("expected error for empty referenceName")
	}
	if err := svc.SubscriptionGroups().Delete(ctx, ""); err == nil {
		t.Error("expected error for empty Delete id")
	}
}
