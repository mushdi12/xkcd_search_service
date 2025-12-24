package core

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
)

type fakeDB struct {
	ids    []int
	idsErr error

	added  []Comics
	addErr error

	stats    DBStats
	statsErr error

	dropErr error
}

func (f *fakeDB) Add(ctx context.Context, c Comics) error {
	f.added = append(f.added, c)
	return f.addErr
}

func (f *fakeDB) Stats(ctx context.Context) (DBStats, error) {
	return f.stats, f.statsErr
}

func (f *fakeDB) Drop(ctx context.Context) error {
	return f.dropErr
}

func (f *fakeDB) IDs(ctx context.Context) ([]int, error) {
	return f.ids, f.idsErr
}

type fakeXKCD struct {
	lastID  int
	lastErr error

	infos  map[int]XKCDInfo
	getErr error
}

func (f fakeXKCD) Get(ctx context.Context, id int) (XKCDInfo, error) {
	if f.getErr != nil {
		return XKCDInfo{}, f.getErr
	}
	return f.infos[id], nil
}

func (f fakeXKCD) LastID(ctx context.Context) (int, error) {
	return f.lastID, f.lastErr
}

type fakeWords struct {
	words []string
	err   error
}

func (f fakeWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	return f.words, f.err
}

type fakeNotificator struct {
	mu     sync.Mutex
	events []EventType
	err    error
}

func (f *fakeNotificator) Publish(ctx context.Context, e EventType) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
	return f.err
}

func newTestService(t *testing.T, db DB, xkcd XKCD, words Words, n Notificator) *Service {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s, err := NewService(logger, db, xkcd, words, 2, "topic", n)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return s
}

func TestNewService_WrongConcurrency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := NewService(logger, &fakeDB{}, fakeXKCD{}, fakeWords{}, 0, "t", &fakeNotificator{}); err == nil {
		t.Fatalf("expected error for concurrency 0")
	}
}

func TestService_Update_Success(t *testing.T) {
	db := &fakeDB{
		ids: []int{1},
	}
	x := fakeXKCD{
		lastID: 3,
		infos: map[int]XKCDInfo{
			2: {ID: 2, URL: "u2", Description: "desc"},
			3: {ID: 3, URL: "u3", Description: "desc"},
		},
	}
	w := fakeWords{words: []string{"w1", "w2"}}
	n := &fakeNotificator{}

	s := newTestService(t, db, x, w, n)

	err := s.Update(context.Background())
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if len(db.added) == 0 {
		t.Fatalf("expected comics to be added")
	}
	if len(n.events) != 1 || n.events[0] != EventTypeUpdating {
		t.Fatalf("expected updating event, got %#v", n.events)
	}
}

func TestService_Update_LockAlreadyHeld(t *testing.T) {
	db := &fakeDB{}
	s := newTestService(t, db, fakeXKCD{}, fakeWords{}, &fakeNotificator{})

	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.Update(context.Background())
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestService_Stats(t *testing.T) {
	db := &fakeDB{
		stats: DBStats{
			WordsTotal:    1,
			WordsUnique:   2,
			ComicsFetched: 3,
		},
	}
	x := fakeXKCD{lastID: 10}

	s := newTestService(t, db, x, fakeWords{}, &fakeNotificator{})

	st, err := s.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}
	if st.ComicsTotal != 10 || st.WordsUnique != 2 {
		t.Fatalf("unexpected stats: %#v", st)
	}
}

func TestService_Status(t *testing.T) {
	s := newTestService(t, &fakeDB{}, fakeXKCD{}, fakeWords{}, &fakeNotificator{})

	if s.Status(context.Background()) != StatusIdle {
		t.Fatalf("expected idle status")
	}
	s.inProgress.Store(true)
	if s.Status(context.Background()) != StatusRunning {
		t.Fatalf("expected running status")
	}
}

func TestService_Drop(t *testing.T) {
	db := &fakeDB{}
	n := &fakeNotificator{}
	s := newTestService(t, db, fakeXKCD{}, fakeWords{}, n)

	if err := s.Drop(context.Background()); err != nil {
		t.Fatalf("Drop returned error: %v", err)
	}
	if len(n.events) != 1 || n.events[0] != EventTypeDropped {
		t.Fatalf("expected dropped event, got %#v", n.events)
	}
}
