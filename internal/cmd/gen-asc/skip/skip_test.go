package skip

import (
	"path/filepath"
	"testing"
)

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
	if got, want := set.Len(), 3; got != want {
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
