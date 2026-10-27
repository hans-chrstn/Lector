package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/user/lector/internal/library"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: path}
	db, err := sql.Open("sqlite3", u.String()+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) migrate() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version > 1 {
		return errors.New("database was created by a newer Lector version")
	}
	var legacy int
	if err := s.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='documents'`).Scan(&legacy); err != nil {
		return err
	}
	if legacy > 0 {
		return errors.New("legacy database detected; use a new DATABASE_PATH to preserve existing data")
	}
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS items (
 id TEXT PRIMARY KEY, title TEXT NOT NULL, author TEXT NOT NULL, kind TEXT NOT NULL,
 status TEXT NOT NULL, current INTEGER NOT NULL, total INTEGER NOT NULL, notes TEXT NOT NULL,
 collection TEXT NOT NULL, file TEXT NOT NULL, format TEXT NOT NULL, created_at TEXT NOT NULL,
 updated_at TEXT NOT NULL, revision INTEGER NOT NULL);
 CREATE INDEX IF NOT EXISTS items_status_updated ON items(status,updated_at DESC);
 CREATE INDEX IF NOT EXISTS items_kind ON items(kind);
 CREATE INDEX IF NOT EXISTS items_collection ON items(collection);
 CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, item_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
 title TEXT NOT NULL, current INTEGER NOT NULL, status TEXT NOT NULL, at TEXT NOT NULL);
 PRAGMA user_version=1;`)
	return err
}

const columns = `id,title,author,kind,status,current,total,notes,collection,file,format,created_at,updated_at,revision`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (library.Item, error) {
	var i library.Item
	var created, updated string
	err := row.Scan(&i.ID, &i.Title, &i.Author, &i.Kind, &i.Status, &i.Current, &i.Total, &i.Notes, &i.Collection, &i.File, &i.Format, &created, &updated, &i.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return i, library.ErrNotFound
	}
	if err != nil {
		return i, err
	}
	i.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return i, err
	}
	i.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	return i, err
}
func (s *Store) Get(ctx context.Context, id string) (library.Item, error) {
	return scan(s.db.QueryRowContext(ctx, `SELECT `+columns+` FROM items WHERE id=?`, id))
}
func (s *Store) List(ctx context.Context, f library.Filter) (library.Page, error) {
	page := library.Page{Items: []library.Item{}}
	where := []string{"1=1"}
	args := []any{}
	if f.Query != "" {
		where = append(where, `(instr(lower(title),lower(?))>0 OR instr(lower(author),lower(?))>0)`)
		args = append(args, f.Query, f.Query)
	}
	for _, pair := range [][2]string{{"kind", f.Kind}, {"status", f.Status}, {"collection", f.Collection}} {
		if pair[1] != "" {
			where = append(where, pair[0]+"=?")
			args = append(args, pair[1])
		}
	}
	clause := strings.Join(where, " AND ")
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM items WHERE `+clause, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	order := "updated_at DESC,id"
	if f.Sort == "title" {
		order = "title COLLATE NOCASE,id"
	}
	if f.Sort == "added" {
		order = "created_at DESC,id"
	}
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 48
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	args = append(args, f.Limit, f.Offset)
	rows, err := s.db.QueryContext(ctx, `SELECT `+columns+` FROM items WHERE `+clause+` ORDER BY `+order+` LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		i, err := scan(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, i)
	}
	return page, rows.Err()
}
func (s *Store) Create(ctx context.Context, i library.Item) (library.Item, error) {
	if err := i.Validate(); err != nil {
		return i, err
	}
	i.ID = rand.Text()
	i.CreatedAt = time.Now().UTC()
	i.UpdatedAt = i.CreatedAt
	i.Revision = 1
	_, err := s.db.ExecContext(ctx, `INSERT INTO items (`+columns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, i.ID, i.Title, i.Author, i.Kind, i.Status, i.Current, i.Total, i.Notes, i.Collection, i.File, i.Format, i.CreatedAt.Format(time.RFC3339Nano), i.UpdatedAt.Format(time.RFC3339Nano), i.Revision)
	return i, err
}
func (s *Store) Update(ctx context.Context, i library.Item) (library.Item, error) {
	if err := i.Validate(); err != nil {
		return i, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return i, err
	}
	defer func() { _ = tx.Rollback() }()
	old, err := scan(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM items WHERE id=?`, i.ID))
	if err != nil {
		return i, err
	}
	if i.Revision != old.Revision {
		return i, library.ErrConflict
	}
	i.File = old.File
	i.Format = old.Format
	i.CreatedAt = old.CreatedAt
	i.UpdatedAt = time.Now().UTC()
	i.Revision++
	_, err = tx.ExecContext(ctx, `UPDATE items SET title=?,author=?,kind=?,status=?,current=?,total=?,notes=?,collection=?,updated_at=?,revision=? WHERE id=?`, i.Title, i.Author, i.Kind, i.Status, i.Current, i.Total, i.Notes, i.Collection, i.UpdatedAt.Format(time.RFC3339Nano), i.Revision, i.ID)
	if err != nil {
		return i, err
	}
	if i.Current != old.Current || i.Status != old.Status {
		_, err = tx.ExecContext(ctx, `INSERT INTO events(item_id,title,current,status,at) VALUES(?,?,?,?,?)`, i.ID, i.Title, i.Current, i.Status, i.UpdatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return i, err
		}
	}
	return i, tx.Commit()
}
func (s *Store) Delete(ctx context.Context, id string) error {
	r, err := s.db.ExecContext(ctx, `DELETE FROM items WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return library.ErrNotFound
	}
	return nil
}
func (s *Store) Stats(ctx context.Context) (library.Stats, error) {
	st := library.Stats{Collections: []string{}}
	err := s.db.QueryRowContext(ctx, `SELECT count(*),coalesce(sum(status='reading'),0),coalesce(sum(status='completed'),0),coalesce(sum(status='planned'),0) FROM items`).Scan(&st.Total, &st.Reading, &st.Completed, &st.Planned)
	if err != nil {
		return st, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT collection FROM items WHERE collection<>'' ORDER BY collection COLLATE NOCASE`)
	if err != nil {
		return st, err
	}
	defer rows.Close()
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return st, err
		}
		st.Collections = append(st.Collections, c)
	}
	return st, rows.Err()
}
func (s *Store) History(ctx context.Context) ([]library.Event, error) {
	result := []library.Event{}
	rows, err := s.db.QueryContext(ctx, `SELECT id,item_id,title,current,status,at FROM events ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e library.Event
		var at string
		if err := rows.Scan(&e.ID, &e.ItemID, &e.Title, &e.Current, &e.Status, &at); err != nil {
			return nil, err
		}
		e.At, err = time.Parse(time.RFC3339Nano, at)
		if err != nil {
			return nil, fmt.Errorf("event time: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
