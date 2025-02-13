package types

import "net/url"

// PaginationToken
// A pagination token that you return to the endpoint on a subsequent call to receive the next set of results.
type PaginationToken string

func (pt PaginationToken) GetParameter() string {
	if len(pt) == 0 {
		return ""
	}
	query := url.Values{}
	query.Set("paginationToken", string(pt))
	if query.Encode() != "" {
		return "?" + query.Encode()
	}
	return query.Encode()

}
