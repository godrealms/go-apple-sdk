package spec

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestFile_NotEmpty(t *testing.T) {
	if len(File) == 0 {
		t.Fatal("spec.File is empty — did `go:embed` work?")
	}
}

func TestSpecSHA256_Matches(t *testing.T) {
	sum := sha256.Sum256(File)
	got := hex.EncodeToString(sum[:])
	if got != SpecSHA256 {
		t.Fatalf("spec SHA256 drift:\n  file = %s\n  const = %s\n"+
			"run scripts/update-spec.sh and update SpecSHA256", got, SpecSHA256)
	}
}

func TestSpecVersion_NonEmpty(t *testing.T) {
	if SpecVersion == "" || SpecSource == "" || SpecDate == "" {
		t.Fatal("SpecVersion/SpecSource/SpecDate must all be set")
	}
}
