package AppStoreConnect

import "encoding/json"

// Document is the top-level JSON:API response envelope.
// It is generic over the primary data attribute type T.
//
// The Data field holds one or many Resource[T] depending on the endpoint.
// For list endpoints, use [Document.AsCollection]; for single-resource
// endpoints, use [Document.AsResource].
//
// Specification: https://jsonapi.org/format/#document-top-level
type Document[T any] struct {
	Data     json.RawMessage `json:"data,omitempty"`
	Included []Resource[any] `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Errors   []Error         `json:"errors,omitempty"`
}

// AsResource decodes Data as a single Resource[T].
// Returns an error if Data is absent or is an array.
func (d *Document[T]) AsResource() (*Resource[T], error) {
	if len(d.Data) == 0 || string(d.Data) == "null" {
		return nil, &ClientError{Message: "document has no data"}
	}
	if d.Data[0] == '[' {
		return nil, &ClientError{Message: "document data is a collection, not a single resource"}
	}
	var r Resource[T]
	if err := json.Unmarshal(d.Data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// AsCollection decodes Data as a slice of Resource[T].
// Returns an error if Data is absent or is an object.
func (d *Document[T]) AsCollection() ([]Resource[T], error) {
	if len(d.Data) == 0 || string(d.Data) == "null" {
		return nil, nil
	}
	if d.Data[0] != '[' {
		return nil, &ClientError{Message: "document data is a single resource, not a collection"}
	}
	var rs []Resource[T]
	if err := json.Unmarshal(d.Data, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

// Resource is a single JSON:API resource object.
// It has a Type discriminator, an Id, typed Attributes, and optional
// Relationships and Links.
//
// Specification: https://jsonapi.org/format/#document-resource-objects
type Resource[T any] struct {
	Type          string                  `json:"type"`
	Id            string                  `json:"id"`
	Attributes    T                       `json:"attributes,omitempty"`
	Relationships map[string]Relationship `json:"relationships,omitempty"`
	Links         *ResourceLinks          `json:"links,omitempty"`
}

// Relationship is a JSON:API relationship object. It may contain linkage
// to other resources via Data, plus navigation links and meta.
//
// For to-one relationships, Data is a single [ResourceIdentifier];
// for to-many, it is an array. Use [Relationship.AsOne] / [Relationship.AsMany]
// to decode.
//
// Specification: https://jsonapi.org/format/#document-resource-object-relationships
type Relationship struct {
	Data  json.RawMessage    `json:"data,omitempty"`
	Links *RelationshipLinks `json:"links,omitempty"`
	Meta  json.RawMessage    `json:"meta,omitempty"`
}

// AsOne decodes Data as a single [ResourceIdentifier].
// Returns (nil, nil) if the relationship has no linkage data.
//
// Uses a value receiver so callers can chain directly from a map index
// expression: rel["foo"].AsOne(). Map index expressions are not
// addressable in Go, so a pointer receiver would force an intermediate
// variable on every call.
func (r Relationship) AsOne() (*ResourceIdentifier, error) {
	if len(r.Data) == 0 || string(r.Data) == "null" {
		return nil, nil
	}
	if r.Data[0] == '[' {
		return nil, &ClientError{Message: "relationship data is a collection, not a single identifier"}
	}
	var id ResourceIdentifier
	if err := json.Unmarshal(r.Data, &id); err != nil {
		return nil, err
	}
	return &id, nil
}

// AsMany decodes Data as a slice of [ResourceIdentifier].
// Returns (nil, nil) if the relationship has no linkage data.
// Uses a value receiver for the same reason as [Relationship.AsOne].
func (r Relationship) AsMany() ([]ResourceIdentifier, error) {
	if len(r.Data) == 0 || string(r.Data) == "null" {
		return nil, nil
	}
	if r.Data[0] != '[' {
		return nil, &ClientError{Message: "relationship data is a single identifier, not a collection"}
	}
	var ids []ResourceIdentifier
	if err := json.Unmarshal(r.Data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// ResourceIdentifier uniquely identifies a JSON:API resource by type and id.
//
// Specification: https://jsonapi.org/format/#document-resource-identifier-objects
type ResourceIdentifier struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

// Links is the top-level links object for a JSON:API document.
// Apple uses "self", "first", "next", and "prev" for pagination.
//
// Specification: https://jsonapi.org/format/#document-links
type Links struct {
	Self  string `json:"self,omitempty"`
	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
}

// ResourceLinks is the links object that appears inside a resource object.
type ResourceLinks struct {
	Self string `json:"self,omitempty"`
}

// RelationshipLinks is the links object that appears inside a relationship.
type RelationshipLinks struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
}
