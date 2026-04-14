package AppStoreConnect

import "context"

// Paginator iterates over pages of a list endpoint, automatically
// following links.next until all pages have been consumed.
//
// Typical usage:
//
//	it := svc.Apps().ListIterator(query)
//	for it.Next(ctx) {
//	    page := it.Page()
//	    for _, app := range page.Data {
//	        handle(app)
//	    }
//	}
//	if err := it.Err(); err != nil {
//	    return err
//	}
//
// Paginator is not safe for concurrent use.
type Paginator[T any] struct {
	svc      *Service
	firstURL string
	query    *Query

	// state
	started  bool
	done     bool
	nextURL  string
	lastPage *Page[T]
	lastErr  error
}

// Page holds a single page of results decoded from a JSON:API document.
type Page[T any] struct {
	Data     []Resource[T]
	Included []Resource[any]
	Links    *Links
}

// newPaginator builds a paginator rooted at path with the given query.
// The initial request uses path + query; subsequent pages follow the
// "next" link verbatim.
func newPaginator[T any](svc *Service, path string, query *Query) *Paginator[T] {
	return &Paginator[T]{
		svc:      svc,
		firstURL: path,
		query:    query,
	}
}

// Next fetches the next page. It returns true if a page was fetched and
// is available via [Paginator.Page], false when there are no more pages
// or an error occurred. Call [Paginator.Err] after Next returns false
// to distinguish end-of-iteration from failure.
func (p *Paginator[T]) Next(ctx context.Context) bool {
	if p.done || p.lastErr != nil {
		return false
	}
	var doc Document[T]
	if !p.started {
		p.started = true
		if _, err := p.svc.do(ctx, "GET", p.firstURL, p.query, nil, &doc); err != nil {
			p.lastErr = err
			return false
		}
	} else {
		if p.nextURL == "" {
			p.done = true
			return false
		}
		if _, err := p.svc.do(ctx, "GET", p.nextURL, nil, nil, &doc); err != nil {
			p.lastErr = err
			return false
		}
	}
	collection, err := doc.AsCollection()
	if err != nil {
		p.lastErr = err
		return false
	}
	p.lastPage = &Page[T]{
		Data:     collection,
		Included: doc.Included,
		Links:    doc.Links,
	}
	if doc.Links != nil && doc.Links.Next != "" {
		p.nextURL = doc.Links.Next
	} else {
		// No more pages after this one. Arm termination for the next
		// Next() call; the caller still needs to read this final page.
		p.nextURL = ""
		p.done = true
	}
	return true
}

// Page returns the most recently fetched page. Only valid after a
// successful call to [Paginator.Next]. Returns nil before the first
// successful call.
func (p *Paginator[T]) Page() *Page[T] { return p.lastPage }

// Err returns the error that stopped iteration, if any.
func (p *Paginator[T]) Err() error { return p.lastErr }

// All consumes the paginator and returns every resource across all pages
// in a single slice. It is a convenience for callers that know their
// result set is small enough to fit comfortably in memory.
func (p *Paginator[T]) All(ctx context.Context) ([]Resource[T], error) {
	var out []Resource[T]
	for p.Next(ctx) {
		out = append(out, p.lastPage.Data...)
	}
	return out, p.lastErr
}
