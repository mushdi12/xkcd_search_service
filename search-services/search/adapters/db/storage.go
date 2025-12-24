package db

import (
	"context"
	"database/sql/driver"
	"log/slog"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/search/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	conn, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}
	return &DB{
		log:  log,
		conn: conn,
	}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

type StringArray []string

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return nil
	}

	str = strings.Trim(str, "{}")
	if str == "" {
		*a = []string{}
		return nil
	}
	parts := strings.Split(str, ",")
	*a = make([]string, len(parts))
	for i, part := range parts {
		(*a)[i] = strings.Trim(part, `"`)
	}
	return nil
}

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	var b strings.Builder
	b.WriteString("{")
	for i, s := range a {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`"`)
		b.WriteString(s)
		b.WriteString(`"`)
	}
	b.WriteString("}")
	return b.String(), nil
}

type Comics struct {
	ID    int         `db:"id"`
	URL   string      `db:"url"`
	Words StringArray `db:"words"`
}

func (db *DB) Search(ctx context.Context, keyword string) ([]int, error) {
	db.log.Info("Search called", "keyword", keyword)
	var IDs []int
	err := db.conn.SelectContext(
		ctx, &IDs,
		"SELECT id FROM comics WHERE $1 = ANY(words)",
		keyword,
	)
	if err != nil {
		db.log.Error("Search query failed", "error", err, "keyword", keyword)
		return nil, err
	}
	db.log.Info("Search results", "count", len(IDs), "keyword", keyword)
	return IDs, err
}

func (db *DB) Get(ctx context.Context, id int) (core.Comics, error) {
	var comics Comics
	err := db.conn.GetContext(
		ctx, &comics,
		"SELECT id, url FROM comics WHERE id = $1",
		id,
	)

	return core.Comics{ID: comics.ID, URL: comics.URL}, err
}

func (db *DB) GetAllComics(ctx context.Context) ([]core.Comics, error) {
	var comics []Comics
	query := `SELECT id, url, words FROM comics`
	err := db.conn.SelectContext(ctx, &comics, query)
	if err != nil {
		return nil, err
	}

	result := make([]core.Comics, len(comics))
	for i, c := range comics {
		result[i] = core.Comics{
			ID:    c.ID,
			URL:   c.URL,
			Words: []string(c.Words),
		}
	}
	return result, nil
}

func (db *DB) GetComicsByIDs(ctx context.Context, ids ...int) ([]core.Comics, error) {
	var comics []Comics
	query := `SELECT id, url, words FROM comics WHERE id = ANY($1::int[])`
	err := db.conn.SelectContext(ctx, &comics, query, ids)
	if err != nil {
		return nil, err
	}

	result := make([]core.Comics, len(comics))
	for i, c := range comics {
		result[i] = core.Comics{
			ID:    c.ID,
			URL:   c.URL,
			Words: []string(c.Words),
		}
	}
	return result, nil
}
