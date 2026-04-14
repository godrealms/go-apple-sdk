package AppStoreConnect

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Error is a single JSON:API error object as returned by the
// App Store Connect API.
//
// Specification: https://jsonapi.org/format/#error-objects
type Error struct {
	Id     string          `json:"id,omitempty"`
	Status string          `json:"status,omitempty"`
	Code   string          `json:"code,omitempty"`
	Title  string          `json:"title,omitempty"`
	Detail string          `json:"detail,omitempty"`
	Source json.RawMessage `json:"source,omitempty"`
	Links  json.RawMessage `json:"links,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`
}

// APIError is returned when the App Store Connect API responds with a
// non-2xx status. It wraps the HTTP status code and all JSON:API errors
// reported in the response body.
//
// Use errors.As to extract it:
//
//	var apiErr *AppStoreConnect.APIError
//	if errors.As(err, &apiErr) {
//	    for _, e := range apiErr.Errors {
//	        if e.Code == "ENTITY_ERROR.RELATIONSHIP.INVALID" { ... }
//	    }
//	}
type APIError struct {
	// StatusCode is the HTTP status code returned by Apple.
	StatusCode int
	// Errors is the parsed list of JSON:API error objects.
	// May be empty if the response body was not a JSON:API error document.
	Errors []Error
	// RawBody is the raw response body, for diagnostics when Errors is empty
	// or when the caller needs to inspect fields the parser does not model.
	RawBody []byte
}

// Error implements the error interface.
func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "app store connect: HTTP %d", e.StatusCode)
	if len(e.Errors) == 0 {
		if len(e.RawBody) > 0 {
			fmt.Fprintf(&b, ": %s", strings.TrimSpace(string(e.RawBody)))
		}
		return b.String()
	}
	for i, err := range e.Errors {
		if i == 0 {
			b.WriteString(": ")
		} else {
			b.WriteString("; ")
		}
		if err.Code != "" {
			fmt.Fprintf(&b, "[%s] ", err.Code)
		}
		if err.Title != "" {
			b.WriteString(err.Title)
		}
		if err.Detail != "" {
			if err.Title != "" {
				b.WriteString(" - ")
			}
			b.WriteString(err.Detail)
		}
	}
	return b.String()
}

// HasCode reports whether any of the underlying errors has the given code.
func (e *APIError) HasCode(code string) bool {
	for _, err := range e.Errors {
		if err.Code == code {
			return true
		}
	}
	return false
}

// ClientError represents an error originating from the SDK itself
// (misconfiguration, local JSON parsing, etc.) as opposed to an error
// returned by Apple. Keeping these distinct from [APIError] lets callers
// tell "Apple rejected the request" from "we never got to Apple".
type ClientError struct {
	Message string
	Cause   error
}

func (e *ClientError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("app store connect: %s: %v", e.Message, e.Cause)
	}
	return "app store connect: " + e.Message
}

func (e *ClientError) Unwrap() error { return e.Cause }

// parseErrorBody attempts to decode body as a JSON:API error document.
// If decoding fails or the body contains no errors, it returns an APIError
// with only StatusCode and RawBody populated.
func parseErrorBody(status int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: status, RawBody: body}
	if len(body) == 0 {
		return apiErr
	}
	var doc struct {
		Errors []Error `json:"errors"`
	}
	if err := json.Unmarshal(body, &doc); err == nil {
		apiErr.Errors = doc.Errors
	}
	return apiErr
}
