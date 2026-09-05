package aether

import (
	"context"
	"fmt"
)

type CursorPage[T any] struct {
	Items      []T
	NextCursor string
}

type PageFetcher[T any] func(context.Context, string) (CursorPage[T], error)

type PagerOptions struct {
	InitialCursor string
	MaxPages      int
}

type PaginationError struct {
	Code string
}

func (err *PaginationError) Error() string {
	switch err.Code {
	case "invalid_max_pages":
		return "maximum pages must be positive"
	case "repeated_cursor":
		return "the server returned a repeated pagination cursor"
	case "max_pages_exceeded":
		return "pagination exceeded the configured page limit"
	default:
		return "pagination failed"
	}
}

type Pager[T any] struct {
	fetcher PageFetcher[T]
	cursor  string
	seen    map[string]struct{}
	pages   int
	max     int
	items   []T
	index   int
	current T
	done    bool
	err     error
}

func NewPager[T any](fetcher PageFetcher[T], supplied ...PagerOptions) *Pager[T] {
	options := PagerOptions{MaxPages: 1_000}
	if len(supplied) == 1 {
		options = supplied[0]
	}
	pager := &Pager[T]{fetcher: fetcher, cursor: options.InitialCursor, seen: make(map[string]struct{}), max: options.MaxPages}
	if fetcher == nil {
		pager.err = fmt.Errorf("aether: page fetcher is required")
	}
	if len(supplied) > 1 {
		pager.err = fmt.Errorf("aether: only one pager options value is allowed")
	}
	if options.MaxPages < 1 {
		pager.err = &PaginationError{Code: "invalid_max_pages"}
	}
	return pager
}

func (pager *Pager[T]) Next(ctx context.Context) bool {
	if pager.err != nil || pager.done {
		return false
	}
	for {
		if pager.index < len(pager.items) {
			pager.current = pager.items[pager.index]
			pager.index++
			return true
		}
		if pager.pages >= pager.max {
			pager.err = &PaginationError{Code: "max_pages_exceeded"}
			return false
		}
		page, err := pager.fetcher(ctx, pager.cursor)
		if err != nil {
			pager.err = err
			return false
		}
		pager.pages++
		pager.items = page.Items
		pager.index = 0
		if page.NextCursor == "" {
			pager.done = true
		} else {
			if page.NextCursor == pager.cursor {
				pager.err = &PaginationError{Code: "repeated_cursor"}
				return false
			}
			if _, exists := pager.seen[page.NextCursor]; exists {
				pager.err = &PaginationError{Code: "repeated_cursor"}
				return false
			}
			if pager.cursor != "" {
				pager.seen[pager.cursor] = struct{}{}
			}
			pager.cursor = page.NextCursor
		}
		if len(pager.items) == 0 && pager.done {
			return false
		}
	}
}

func (pager *Pager[T]) Value() T {
	return pager.current
}

func (pager *Pager[T]) Err() error {
	return pager.err
}
