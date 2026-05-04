package jws

import (
	"encoding/base64"
	"testing"

	"github.com/godrealms/go-apple-sdk/internal/testchain"
)

func TestX5c_Parse_Success(t *testing.T) {
	tc := testchain.New(t)
	x := X5c{
		base64.StdEncoding.EncodeToString(tc.Leaf.Raw),
		base64.StdEncoding.EncodeToString(tc.Intermediate.Raw),
		base64.StdEncoding.EncodeToString(tc.Root.Raw),
	}
	certs, err := x.Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(certs) != 3 {
		t.Fatalf("expected 3 certs, got %d", len(certs))
	}
	if !certs[0].Equal(tc.Leaf) {
		t.Fatalf("certs[0] != leaf")
	}
}

func TestX5c_Parse_EmptyChain(t *testing.T) {
	var x X5c
	_, err := x.Parse()
	assertReason(t, err, ReasonStructure)
}

func TestX5c_Parse_BadBase64(t *testing.T) {
	x := X5c{"!!!!not-base64!!!!"}
	_, err := x.Parse()
	assertReason(t, err, ReasonStructure)
}

func TestX5c_Parse_BadDER(t *testing.T) {
	x := X5c{base64.StdEncoding.EncodeToString([]byte("not a cert"))}
	_, err := x.Parse()
	assertReason(t, err, ReasonStructure)
}
