package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"

	"yadro.com/course/update/core"
)

type fakeSQLXDB struct {
	execErr      error
	getErr       error
	selectErr    error
	closeErr     error
	getCallCount int
	getResults   []interface{}
	selectResult interface{}
}

func (f *fakeSQLXDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, f.execErr
}

func (f *fakeSQLXDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if f.getErr != nil {
		return f.getErr
	}
	if f.getResults != nil && f.getCallCount < len(f.getResults) {
		result := f.getResults[f.getCallCount]
		f.getCallCount++
		if d, ok := dest.(*int); ok {
			*d = result.(int)
		} else if d, ok := dest.(*core.DBStats); ok {
			*d = result.(core.DBStats)
		}
	}
	return nil
}

func (f *fakeSQLXDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if f.selectErr != nil {
		return f.selectErr
	}
	if f.selectResult != nil {
		if d, ok := dest.(*[]int); ok {
			*d = f.selectResult.([]int)
		}
	}
	return nil
}

func (f *fakeSQLXDB) Close() error {
	return f.closeErr
}

func TestDB_Add_Success(t *testing.T) {
	fakeConn := &fakeSQLXDB{}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	err := db.Add(context.Background(), core.Comics{
		ID:    1,
		URL:   "http://example.com",
		Words: []string{"test"},
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
}

func TestDB_Add_Error(t *testing.T) {
	fakeConn := &fakeSQLXDB{execErr: errors.New("db error")}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	err := db.Add(context.Background(), core.Comics{ID: 1})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestDB_Stats_Success(t *testing.T) {
	fakeConn := &fakeSQLXDB{
		getResults: []interface{}{10, 100, 50},
	}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	stats, err := db.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}
	if stats.ComicsFetched != 10 {
		t.Fatalf("expected ComicsFetched=10, got %d", stats.ComicsFetched)
	}
	if stats.WordsTotal != 100 {
		t.Fatalf("expected WordsTotal=100, got %d", stats.WordsTotal)
	}
	if stats.WordsUnique != 50 {
		t.Fatalf("expected WordsUnique=50, got %d", stats.WordsUnique)
	}
}

func TestDB_Stats_Error(t *testing.T) {
	fakeConn := &fakeSQLXDB{getErr: errors.New("db error")}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	_, err := db.Stats(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestDB_IDs_Success(t *testing.T) {
	fakeConn := &fakeSQLXDB{
		selectResult: []int{1, 2, 3},
	}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	ids, err := db.IDs(context.Background())
	if err != nil {
		t.Fatalf("IDs returned error: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("expected 3 ids, got %d", len(ids))
	}
}

func TestDB_IDs_NoRows(t *testing.T) {
	fakeConn := &fakeSQLXDB{selectErr: sql.ErrNoRows}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	ids, err := db.IDs(context.Background())
	if err != nil {
		t.Fatalf("IDs returned error: %v", err)
	}
	if ids != nil {
		t.Fatalf("expected nil ids, got %v", ids)
	}
}

func TestDB_Drop_Success(t *testing.T) {
	fakeConn := &fakeSQLXDB{}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	err := db.Drop(context.Background())
	if err != nil {
		t.Fatalf("Drop returned error: %v", err)
	}
}

func TestDB_Close(t *testing.T) {
	fakeConn := &fakeSQLXDB{}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	err := db.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestDB_Migrate_RequiresRealDB(t *testing.T) {
	fakeConn := &fakeSQLXDB{}
	db := &DB{
		log:  slog.Default(),
		conn: fakeConn,
	}

	err := db.Migrate()
	if err == nil {
		t.Fatalf("expected error for fake connection, got nil")
	}
}
