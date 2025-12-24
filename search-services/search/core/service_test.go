package core

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type fakeStorager struct {
	searchResults map[string][]int
	comics        map[int]Comics

	searchErr error
	getErrID  int
	getErr    error
}

func (f fakeStorager) Search(ctx context.Context, keyword string) ([]int, error) {
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	return f.searchResults[keyword], nil
}

func (f fakeStorager) Get(ctx context.Context, id int) (Comics, error) {
	if f.getErrID != 0 && id == f.getErrID {
		return Comics{}, f.getErr
	}
	return f.comics[id], nil
}

func (f fakeStorager) GetAllComics(ctx context.Context) ([]Comics, error) {
	return nil, nil
}

func (f fakeStorager) GetComicsByIDs(ctx context.Context, ids ...int) ([]Comics, error) {
	return nil, nil
}

type fakeWords struct {
	words []string
	err   error
}

func (f fakeWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	return f.words, f.err
}

type fakeInitiator struct {
	indexedComics []Comics
	err           error
}

func (f fakeInitiator) GetIndexedComics(ctx context.Context, words []string, limit int) ([]Comics, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.indexedComics) < limit {
		limit = len(f.indexedComics)
	}
	return f.indexedComics[:limit], nil
}

func (f fakeInitiator) IndexComics(ctx context.Context) error {
	return nil
}

func (f fakeInitiator) ClearIndex(ctx context.Context) error {
	return nil
}

func newTestService(t *testing.T, db Storager, w Words, init Initiator) *Service {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	s, err := NewService(logger, db, w, init)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return s
}

func TestService_Search_Success(t *testing.T) {
	ctx := context.Background()

	db := fakeStorager{
		searchResults: map[string][]int{
			"linux": {1, 2},
			"cpu":   {2, 3},
		},
		comics: map[int]Comics{
			1: {ID: 1, URL: "url1"},
			2: {ID: 2, URL: "url2"},
			3: {ID: 3, URL: "url3"},
		},
	}

	words := fakeWords{
		words: []string{"linux", "cpu"},
	}

	s := newTestService(t, db, words, fakeInitiator{})

	result, err := s.Search(ctx, "linux cpu", 2)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 comics, got %d", len(result))
	}

	if result[0].ID != 2 {
		t.Fatalf("expected first result to have ID=2, got %d", result[0].ID)
	}
}

func TestService_Search_WordsError(t *testing.T) {
	ctx := context.Background()

	db := fakeStorager{}
	words := fakeWords{
		err: errors.New("norm error"),
	}

	s := newTestService(t, db, words, fakeInitiator{})

	result, err := s.Search(ctx, "phrase", 10)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result on error, got %#v", result)
	}
}

func TestService_IndexSearch_Success(t *testing.T) {
	ctx := context.Background()

	words := fakeWords{
		words: []string{"linux"},
	}

	init := fakeInitiator{
		indexedComics: []Comics{
			{ID: 1, URL: "url1"},
			{ID: 2, URL: "url2"},
		},
	}

	s := newTestService(t, fakeStorager{}, words, init)

	result, err := s.IndexSearch(ctx, "linux", 1)
	if err != nil {
		t.Fatalf("IndexSearch returned error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 comics, got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Fatalf("expected comics ID=1, got %d", result[0].ID)
	}
}
