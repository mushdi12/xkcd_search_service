package db

import (
	"context"
	"database/sql"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
)

type sqlxDB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Close() error
}

type DB struct {
	log  *slog.Logger
	conn sqlxDB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	_, err := db.conn.ExecContext(
		ctx,
		"INSERT INTO comics (id, url, words) VALUES ($1, $2, $3)",
		comics.ID, comics.URL, comics.Words)
	return err
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats

	err := db.conn.GetContext(ctx, &stats.ComicsFetched, "SELECT COUNT(*) FROM comics")
	if err != nil {
		return core.DBStats{}, err
	}

	err = db.conn.GetContext(ctx, &stats.WordsTotal, "SELECT coalesce(SUM(array_length(words,1)), 0) FROM comics")
	if err != nil {
		return core.DBStats{}, err
	}

	err = db.conn.GetContext(ctx, &stats.WordsUnique, "SELECT COUNT(DISTINCT word) FROM comics, unnest(words) AS word")
	if err != nil {
		return core.DBStats{}, err
	}

	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	err := db.conn.SelectContext(ctx, &ids, "SELECT id FROM comics")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, "TRUNCATE comics")
	return err
}
