package AppStoreConnect

import (
	"context"
	"net/http"
	"testing"
)

// Tests in this file exercise the List / Get / Delete code paths on
// the services introduced in Phase 3 that aren't covered by the
// higher-level happy-path tests elsewhere. They use a single
// dispatcher so each service only needs the minimum fixture to prove
// the endpoint wiring works; deeper business assertions live in
// phase3_test.go.

// minimalDoc returns a JSON document with a single resource of the
// given type and id, and empty attributes.
func minimalDoc(typ, id string) string {
	return `{"data":{"type":"` + typ + `","id":"` + id + `","attributes":{}}}`
}

// minimalCollectionDoc returns a JSON document containing an empty
// data collection. Sufficient to prove a List endpoint decoded OK.
func minimalCollectionDoc() string {
	return `{"data":[]}`
}

func TestPhase3_ListGetDeletePaths(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		// Collection GETs
		case r.Method == "GET" && (r.URL.Path == "/v1/betaGroups" ||
			r.URL.Path == "/v1/bundleIds" ||
			r.URL.Path == "/v1/certificates" ||
			r.URL.Path == "/v1/profiles" ||
			r.URL.Path == "/v1/userInvitations"):
			_, _ = w.Write([]byte(minimalCollectionDoc()))

		// Resource GETs
		case r.Method == "GET" && r.URL.Path == "/v1/betaGroups/bg-1":
			_, _ = w.Write([]byte(minimalDoc("betaGroups", "bg-1")))
		case r.Method == "GET" && r.URL.Path == "/v1/builds/build-1":
			_, _ = w.Write([]byte(minimalDoc("builds", "build-1")))
		case r.Method == "GET" && r.URL.Path == "/v1/bundleIds/bid-1":
			_, _ = w.Write([]byte(minimalDoc("bundleIds", "bid-1")))
		case r.Method == "GET" && r.URL.Path == "/v1/certificates/cert-1":
			_, _ = w.Write([]byte(minimalDoc("certificates", "cert-1")))
		case r.Method == "GET" && r.URL.Path == "/v1/profiles/prof-1":
			_, _ = w.Write([]byte(minimalDoc("profiles", "prof-1")))
		case r.Method == "GET" && r.URL.Path == "/v1/users/user-1":
			_, _ = w.Write([]byte(minimalDoc("users", "user-1")))

		// Deletes — all return 204 with no body.
		case r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)

		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))

	ctx := context.Background()

	// Collection GETs
	if _, err := svc.BetaGroups().List(ctx, nil); err != nil {
		t.Errorf("BetaGroups.List: %v", err)
	}
	if _, err := svc.BundleIDs().List(ctx, nil); err != nil {
		t.Errorf("BundleIDs.List: %v", err)
	}
	if _, err := svc.Certificates().List(ctx, nil); err != nil {
		t.Errorf("Certificates.List: %v", err)
	}
	if _, err := svc.Profiles().List(ctx, nil); err != nil {
		t.Errorf("Profiles.List: %v", err)
	}
	if _, err := svc.UserInvitations().List(ctx, nil); err != nil {
		t.Errorf("UserInvitations.List: %v", err)
	}

	// Resource GETs
	if _, err := svc.BetaGroups().Get(ctx, "bg-1", nil); err != nil {
		t.Errorf("BetaGroups.Get: %v", err)
	}
	if _, err := svc.Builds().Get(ctx, "build-1", nil); err != nil {
		t.Errorf("Builds.Get: %v", err)
	}
	if _, err := svc.BundleIDs().Get(ctx, "bid-1", nil); err != nil {
		t.Errorf("BundleIDs.Get: %v", err)
	}
	if _, err := svc.Certificates().Get(ctx, "cert-1", nil); err != nil {
		t.Errorf("Certificates.Get: %v", err)
	}
	if _, err := svc.Profiles().Get(ctx, "prof-1", nil); err != nil {
		t.Errorf("Profiles.Get: %v", err)
	}
	if _, err := svc.Users().Get(ctx, "user-1", nil); err != nil {
		t.Errorf("Users.Get: %v", err)
	}

	// Deletes
	if err := svc.BundleIDs().Delete(ctx, "bid-1"); err != nil {
		t.Errorf("BundleIDs.Delete: %v", err)
	}
	if err := svc.Certificates().Delete(ctx, "cert-1"); err != nil {
		t.Errorf("Certificates.Delete: %v", err)
	}
	if err := svc.Profiles().Delete(ctx, "prof-1"); err != nil {
		t.Errorf("Profiles.Delete: %v", err)
	}
	if err := svc.UserInvitations().Delete(ctx, "inv-1"); err != nil {
		t.Errorf("UserInvitations.Delete: %v", err)
	}

	// Iterator accessors — just verify they return non-nil paginators.
	// Calling Next() on a dispatch-backed server would require
	// per-method fixtures, so we stop short of driving them; the
	// iterator code path itself is already covered via apps_test.go
	// and customer_reviews_test.go.
	if svc.Builds().ListIterator(nil) == nil {
		t.Error("Builds.ListIterator nil")
	}
	if svc.BetaGroups().ListIterator(nil) == nil {
		t.Error("BetaGroups.ListIterator nil")
	}
	if svc.BundleIDs().ListIterator(nil) == nil {
		t.Error("BundleIDs.ListIterator nil")
	}
	if svc.Certificates().ListIterator(nil) == nil {
		t.Error("Certificates.ListIterator nil")
	}
	if svc.Profiles().ListIterator(nil) == nil {
		t.Error("Profiles.ListIterator nil")
	}
	if svc.UserInvitations().ListIterator(nil) == nil {
		t.Error("UserInvitations.ListIterator nil")
	}
}

func TestBuildsService_ListAllSmoke(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "builds_list.json"))
	}))
	all, err := svc.Builds().ListAll(context.Background(), NewQuery().Limit(2))
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("len = %d", len(all))
	}
}
