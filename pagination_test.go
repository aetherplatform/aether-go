package aether

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPagerTraversesPages(t *testing.T) {
	t.Parallel()
	pages := map[string]CursorPage[int]{
		"":     {Items: []int{1, 2}, NextCursor: "next"},
		"next": {Items: []int{3}},
	}
	pager := NewPager(func(_ context.Context, cursor string) (CursorPage[int], error) {
		return pages[cursor], nil
	})
	var values []int
	for pager.Next(context.Background()) {
		values = append(values, pager.Value())
	}
	if err := pager.Err(); err != nil {
		t.Fatal(err)
	}
	if len(values) != 3 || values[0] != 1 || values[2] != 3 {
		t.Fatalf("unexpected values: %v", values)
	}
}

func TestPagerRejectsRepeatedCursor(t *testing.T) {
	t.Parallel()
	pager := NewPager(func(_ context.Context, _ string) (CursorPage[int], error) {
		return CursorPage[int]{NextCursor: "same"}, nil
	})
	if pager.Next(context.Background()) {
		t.Fatal("expected pagination to stop")
	}
	var paginationErr *PaginationError
	if !errors.As(pager.Err(), &paginationErr) || paginationErr.Code != "repeated_cursor" {
		t.Fatalf("unexpected error: %v", pager.Err())
	}
}

func TestPagerRejectsMultipleOptions(t *testing.T) {
	t.Parallel()
	pager := NewPager(func(_ context.Context, _ string) (CursorPage[int], error) {
		return CursorPage[int]{}, nil
	}, PagerOptions{MaxPages: 1}, PagerOptions{MaxPages: 2})
	if pager.Next(context.Background()) {
		t.Fatal("expected pagination to stop")
	}
	if got := pager.Err(); got == nil || got.Error() != "aether: only one pager options value is allowed" {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestPaginationVectors(t *testing.T) {
	t.Parallel()
	var document struct {
		Cases []struct {
			Name     string    `json:"name"`
			Cursors  []*string `json:"cursors"`
			MaxPages int       `json:"max_pages"`
			Expected string    `json:"expected"`
		} `json:"cases"`
	}
	paths := []string{
		filepath.Join("..", "..", "contracts", "sdk", "v1", "pagination-vectors.json"),
		filepath.Join("contracts", "sdk", "v1", "pagination-vectors.json"),
	}
	var data []byte
	var err error
	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range document.Cases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			index := 0
			pager := NewPager(func(_ context.Context, _ string) (CursorPage[int], error) {
				if index >= len(testCase.Cursors) {
					return CursorPage[int]{}, nil
				}
				next := ""
				if testCase.Cursors[index] != nil {
					next = *testCase.Cursors[index]
				}
				index++
				return CursorPage[int]{Items: []int{index}, NextCursor: next}, nil
			}, PagerOptions{MaxPages: testCase.MaxPages})
			for pager.Next(context.Background()) {
			}
			if testCase.Expected == "complete" {
				if pager.Err() != nil {
					t.Fatal(pager.Err())
				}
				return
			}
			var paginationErr *PaginationError
			if !errors.As(pager.Err(), &paginationErr) || paginationErr.Code != testCase.Expected {
				t.Fatalf("error=%v, want %s", pager.Err(), testCase.Expected)
			}
		})
	}
}
