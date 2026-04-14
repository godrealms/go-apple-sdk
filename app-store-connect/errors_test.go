package AppStoreConnect

import (
	"errors"
	"strings"
	"testing"
)

func TestAPIError_ErrorFormatting(t *testing.T) {
	e := &APIError{
		StatusCode: 403,
		Errors: []Error{
			{Code: "FORBIDDEN_ERROR", Title: "Nope", Detail: "API key not authorized"},
		},
	}
	msg := e.Error()
	if !strings.Contains(msg, "HTTP 403") {
		t.Errorf("missing status: %q", msg)
	}
	if !strings.Contains(msg, "FORBIDDEN_ERROR") {
		t.Errorf("missing code: %q", msg)
	}
	if !strings.Contains(msg, "Nope") || !strings.Contains(msg, "API key not authorized") {
		t.Errorf("missing title/detail: %q", msg)
	}
}

func TestAPIError_EmptyErrorsFallsBackToRawBody(t *testing.T) {
	e := &APIError{
		StatusCode: 500,
		RawBody:    []byte("internal server error\n"),
	}
	msg := e.Error()
	if !strings.Contains(msg, "HTTP 500") {
		t.Errorf("missing status: %q", msg)
	}
	if !strings.Contains(msg, "internal server error") {
		t.Errorf("missing raw body: %q", msg)
	}
}

func TestAPIError_HasCode(t *testing.T) {
	e := &APIError{
		Errors: []Error{
			{Code: "ENTITY_ERROR.RELATIONSHIP.INVALID"},
			{Code: "FORBIDDEN_ERROR"},
		},
	}
	if !e.HasCode("FORBIDDEN_ERROR") {
		t.Error("expected HasCode true for present code")
	}
	if e.HasCode("NOT_PRESENT") {
		t.Error("expected HasCode false for absent code")
	}
}

func TestParseErrorBody_ValidJSONAPIError(t *testing.T) {
	body := []byte(`{"errors":[{"status":"409","code":"STATE_ERROR","title":"Conflict","detail":"Bad state"}]}`)
	e := parseErrorBody(409, body)
	if e.StatusCode != 409 {
		t.Errorf("StatusCode = %d", e.StatusCode)
	}
	if len(e.Errors) != 1 {
		t.Fatalf("Errors len = %d, want 1", len(e.Errors))
	}
	if e.Errors[0].Code != "STATE_ERROR" {
		t.Errorf("code = %q", e.Errors[0].Code)
	}
}

func TestParseErrorBody_NotJSON(t *testing.T) {
	body := []byte("<html>gateway timeout</html>")
	e := parseErrorBody(504, body)
	if e.StatusCode != 504 {
		t.Errorf("StatusCode = %d", e.StatusCode)
	}
	if len(e.Errors) != 0 {
		t.Errorf("Errors should be empty, got %d", len(e.Errors))
	}
	if string(e.RawBody) != string(body) {
		t.Errorf("RawBody lost")
	}
}

func TestClientError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	ce := &ClientError{Message: "outer", Cause: inner}
	if !errors.Is(ce, inner) {
		t.Error("errors.Is should traverse ClientError.Cause")
	}
	if !strings.Contains(ce.Error(), "outer") || !strings.Contains(ce.Error(), "inner") {
		t.Errorf("error format: %q", ce.Error())
	}
}
