package initiator

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"yadro.com/course/search/core"
)

type fakeDB struct {
	allComics      []core.Comics
	allComicsErr   error
	comicsByIDs    []core.Comics
	comicsByIDsErr error
	lastGetArgs    []int
}

func (f *fakeDB) Search(ctx context.Context, keyword string) ([]int, error) {
	return nil, nil
}

func (f *fakeDB) Get(ctx context.Context, id int) (core.Comics, error) {
	return core.Comics{}, nil
}

func (f *fakeDB) GetAllComics(ctx context.Context) ([]core.Comics, error) {
	return f.allComics, f.allComicsErr
}

func (f *fakeDB) GetComicsByIDs(ctx context.Context, ids ...int) ([]core.Comics, error) {
	f.lastGetArgs = ids

	if f.comicsByIDsErr != nil {
		return nil, f.comicsByIDsErr
	}
	if len(f.comicsByIDs) < len(ids) {
		return f.comicsByIDs, nil
	}
	return f.comicsByIDs[:len(ids)], nil
}

func newTestInitiator(db core.Storager) *Initiator {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewInitiator(logger, db, time.Minute)
}

func TestInitiator_IndexComics_BuildsIndex(t *testing.T) {
	db := &fakeDB{
		allComics: []core.Comics{
			{ID: 1, URL: "u1", Words: []string{"a", "b"}},
			{ID: 2, URL: "u2", Words: []string{"b", "c"}},
		},
	}
	init := newTestInitiator(db)

	err := init.IndexComics(context.Background())
	if err != nil {
		t.Fatalf("IndexComics returned error: %v", err)
	}

	init.mu.RLock()
	defer init.mu.RUnlock()

	if len(init.indexedComics) != 2 {
		t.Fatalf("expected 2 indexed comics, got %d", len(init.indexedComics))
	}
}

func TestInitiator_GetIndexedComics_EmptyWords(t *testing.T) {
	init := newTestInitiator(&fakeDB{})

	res, err := init.GetIndexedComics(context.Background(), nil, 10)
	if err != nil {
		t.Fatalf("GetIndexedComics returned error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected empty result, got %d", len(res))
	}
}

func TestInitiator_GetIndexedComics_ScoringAndLimit(t *testing.T) {
	db := &fakeDB{
		comicsByIDs: []core.Comics{
			{ID: 1, URL: "u1"},
			{ID: 2, URL: "u2"},
		},
	}
	init := newTestInitiator(db)

	init.indexedComics["1"] = []string{"linux", "cpu"}
	init.indexedComics["2"] = []string{"linux"}

	res, err := init.GetIndexedComics(context.Background(), []string{"linux", "cpu"}, 1)
	if err != nil {
		t.Fatalf("GetIndexedComics returned error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].ID != 1 {
		t.Fatalf("expected id=1, got %d", res[0].ID)
	}
}

func TestInitiator_ClearIndex(t *testing.T) {
	init := newTestInitiator(&fakeDB{})
	init.indexedComics["1"] = []string{"a"}

	err := init.ClearIndex(context.Background())
	if err != nil {
		t.Fatalf("ClearIndex returned error: %v", err)
	}

	init.mu.RLock()
	defer init.mu.RUnlock()

	if len(init.indexedComics) != 0 {
		t.Fatalf("expected empty index, got %d", len(init.indexedComics))
	}
}
