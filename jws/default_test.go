package jws

import (
	"sync"
	"testing"
)

func TestDefaultVerifier_LoadsEmbeddedRoot(t *testing.T) {
	v := DefaultVerifier()
	if v == nil {
		t.Fatalf("DefaultVerifier returned nil")
	}
	if v.roots == nil {
		t.Fatalf("DefaultVerifier root pool is nil")
	}
	if len(v.roots.Subjects()) == 0 { //nolint:staticcheck // Subjects is fine for a sanity check
		t.Fatalf("DefaultVerifier root pool is empty")
	}
}

func TestDefaultVerifier_Singleton(t *testing.T) {
	a := DefaultVerifier()
	b := DefaultVerifier()
	if a != b {
		t.Fatalf("DefaultVerifier should return the same instance across calls")
	}
}

func TestDefaultVerifier_ConcurrentAccess(t *testing.T) {
	const n = 100
	var wg sync.WaitGroup
	verifiers := make([]*Verifier, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			verifiers[i] = DefaultVerifier()
		}(i)
	}
	wg.Wait()
	for i := 1; i < n; i++ {
		if verifiers[i] != verifiers[0] {
			t.Fatalf("verifier[%d] differs from verifier[0]", i)
		}
	}
}

func TestDefaultVerifier_UsesDefaultRequiredOIDs(t *testing.T) {
	v := DefaultVerifier()
	if len(v.requiredOIDs) != len(DefaultRequiredOIDs) {
		t.Fatalf("DefaultVerifier OID list = %d, want %d", len(v.requiredOIDs), len(DefaultRequiredOIDs))
	}
	for i, oid := range v.requiredOIDs {
		if !oid.Equal(DefaultRequiredOIDs[i]) {
			t.Fatalf("DefaultVerifier OID[%d] = %v, want %v", i, oid, DefaultRequiredOIDs[i])
		}
	}
}
