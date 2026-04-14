package AppStoreConnect

import (
	"encoding/json"
	"testing"
)

func TestDocument_AsResource(t *testing.T) {
	body := loadFixture(t, "app_get.json")
	var doc Document[AppAttributes]
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	res, err := doc.AsResource()
	if err != nil {
		t.Fatalf("AsResource: %v", err)
	}
	if res.Id != "1234567890" {
		t.Errorf("Id = %q", res.Id)
	}
	if res.Type != "apps" {
		t.Errorf("Type = %q", res.Type)
	}
	if res.Attributes.BundleId != "com.acme.widgets" {
		t.Errorf("BundleId = %q", res.Attributes.BundleId)
	}
	// Verify Relationships decode and the to-one/to-many helpers work.
	versions, ok := res.Relationships["appStoreVersions"]
	if !ok {
		t.Fatal("missing appStoreVersions relationship")
	}
	many, err := versions.AsMany()
	if err != nil {
		t.Fatalf("AsMany: %v", err)
	}
	if len(many) != 2 {
		t.Errorf("versions len = %d, want 2", len(many))
	}
	if many[0].Type != "appStoreVersions" || many[0].Id != "v-1" {
		t.Errorf("version[0] = %+v", many[0])
	}

	ci, ok := res.Relationships["ciProduct"]
	if !ok {
		t.Fatal("missing ciProduct relationship")
	}
	one, err := ci.AsOne()
	if err != nil {
		t.Fatalf("AsOne: %v", err)
	}
	if one == nil || one.Id != "ci-xyz" {
		t.Errorf("ci = %+v", one)
	}
}

func TestDocument_AsCollection(t *testing.T) {
	body := loadFixture(t, "apps_list_page1.json")
	var doc Document[AppAttributes]
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	col, err := doc.AsCollection()
	if err != nil {
		t.Fatalf("AsCollection: %v", err)
	}
	if len(col) != 2 {
		t.Fatalf("len = %d, want 2", len(col))
	}
	if col[0].Attributes.Name != "Acme Widgets" {
		t.Errorf("col[0].Name = %q", col[0].Attributes.Name)
	}
	if doc.Links == nil || doc.Links.Next == "" {
		t.Error("expected links.next to be set")
	}
}

func TestDocument_AsResourceOnCollectionFails(t *testing.T) {
	body := loadFixture(t, "apps_list_page1.json")
	var doc Document[AppAttributes]
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, err := doc.AsResource(); err == nil {
		t.Error("expected error when calling AsResource on collection data")
	}
}

func TestDocument_AsCollectionOnSingleFails(t *testing.T) {
	body := loadFixture(t, "app_get.json")
	var doc Document[AppAttributes]
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, err := doc.AsCollection(); err == nil {
		t.Error("expected error when calling AsCollection on single resource data")
	}
}

func TestRelationship_EmptyDataReturnsNil(t *testing.T) {
	r := Relationship{}
	one, err := r.AsOne()
	if err != nil {
		t.Errorf("AsOne empty: %v", err)
	}
	if one != nil {
		t.Errorf("expected nil, got %+v", one)
	}
	many, err := r.AsMany()
	if err != nil {
		t.Errorf("AsMany empty: %v", err)
	}
	if many != nil {
		t.Errorf("expected nil, got %+v", many)
	}
}
