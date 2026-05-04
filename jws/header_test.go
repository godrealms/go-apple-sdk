package jws

import (
	"encoding/json"
	"testing"
)

func TestHeader_Unmarshal(t *testing.T) {
	raw := `{"alg":"ES256","x5c":["abc","def"]}`
	var h Header
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if h.Alg != "ES256" {
		t.Fatalf("Alg = %q, want ES256", h.Alg)
	}
	if len(h.X5c) != 2 || h.X5c[0] != "abc" {
		t.Fatalf("X5c = %v, want [abc def]", h.X5c)
	}
}

func TestHeader_UnmarshalIgnoresUnknownFields(t *testing.T) {
	// Apple may include kid, typ, etc. — encoding/json should
	// silently ignore them.
	raw := `{"alg":"ES256","x5c":["abc"],"kid":"1234","typ":"JWT"}`
	var h Header
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("unmarshal with extras: %v", err)
	}
	if h.Alg != "ES256" {
		t.Fatalf("Alg = %q, want ES256", h.Alg)
	}
}
