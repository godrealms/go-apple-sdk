package skip

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFileImpl wraps os.WriteFile for use by the tiny writeFile
// helper at the bottom of this file. Kept private so the test
// imports stay tidy.
func writeFileImpl(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func TestLoad_Fixture(t *testing.T) {
	set, err := Load(filepath.Join("testdata", "fixture.txt"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, name := range []string{"apps", "builds", "customerReviews"} {
		if !set.Contains(name) {
			t.Errorf("Contains(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"widgets", "appsExtra", ""} {
		if set.Contains(name) {
			t.Errorf("Contains(%q) = true, want false", name)
		}
	}
	// fixture.txt has 3 resource-level lines (apps appears twice — collapsed)
	// + 3 per-path lines under appStoreVersions. Total = 6.
	if got, want := set.Len(), 6; got != want {
		t.Errorf("Len = %d, want %d", got, want)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("testdata/does_not_exist.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_IgnoresBlankAndComments(t *testing.T) {
	set, err := Load(filepath.Join("testdata", "fixture.txt"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// fixture.txt explicitly exercises blanks, comments, and trailing
	// whitespace; they must not pollute the set.
	if set.Contains("# comment") || set.Contains("  apps  ") {
		t.Error("raw comment or untrimmed line leaked into set")
	}
}

func TestSkip_ResourceLevel(t *testing.T) {
	set, err := Load(filepath.Join("testdata", "fixture.txt"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Resource-level skips fire regardless of path.
	if !set.Skip("apps", "/v1/apps") {
		t.Errorf("Skip(apps, /v1/apps) = false; want true (resource-level)")
	}
	if !set.Skip("apps", "/v1/apps/{id}/customerReviews") {
		t.Errorf("Skip(apps, sub-path) = false; want true (resource-level)")
	}
}

func TestSkip_ExactPath(t *testing.T) {
	set, err := Load(filepath.Join("testdata", "fixture.txt"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// appStoreVersions has per-path entries; resource itself is NOT
	// in the resource-level set, so non-matching paths must pass.
	if set.Contains("appStoreVersions") {
		t.Fatalf("appStoreVersions should be per-path only, not resource-level")
	}
	if !set.Skip("appStoreVersions", "/v1/appStoreVersions") {
		t.Errorf("Skip(appStoreVersions, /v1/appStoreVersions) = false; want true (exact match)")
	}
	if !set.Skip("appStoreVersions", "/v1/appStoreVersions/{id}") {
		t.Errorf("Skip(appStoreVersions, /v1/appStoreVersions/{id}) = false; want true (exact match)")
	}
	if set.Skip("appStoreVersions", "/v1/appStoreVersions/{id}/build") {
		t.Errorf("Skip(appStoreVersions, /build) = true; want false (no exact match, no prefix match)")
	}
}

func TestSkip_PrefixGlob(t *testing.T) {
	set, err := Load(filepath.Join("testdata", "fixture.txt"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// "appStoreVersions /v1/appStoreVersions/{id}/relationships/*"
	prefix := "/v1/appStoreVersions/{id}/relationships"
	for _, path := range []string{
		prefix,                // pattern itself
		prefix + "/build",     // direct child
		prefix + "/appStoreVersionLocalizations",
	} {
		if !set.Skip("appStoreVersions", path) {
			t.Errorf("Skip(appStoreVersions, %q) = false; want true (prefix glob)", path)
		}
	}
	// Sibling path NOT under prefix must pass.
	if set.Skip("appStoreVersions", "/v1/appStoreVersions/{id}/anotherSubresource") {
		t.Errorf("prefix glob over-matched")
	}
}

func TestLoad_RejectsThreeTokens(t *testing.T) {
	// Format check: 3+ tokens on a non-comment line must error.
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.txt")
	if err := writeFile(path, "apps GET /v1/apps\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error on 3-token line")
	}
}

// writeFile is a tiny helper to keep the test free of os.WriteFile
// imports cluttering the file header.
func writeFile(path, content string) error {
	return writeFileImpl(path, content)
}
