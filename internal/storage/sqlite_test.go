package storage

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/user/lector/internal/library"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestRelativeDatabasePath(t *testing.T) {
	t.Chdir(t.TempDir())
	s, err := Open("data/library.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.Create(context.Background(), library.Item{Title: "Local", Kind: "book", Status: "planned"}); err != nil {
		t.Fatal(err)
	}
}
func TestPersistentLibraryAndHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	item, err := s.Create(ctx, library.Item{Title: " A quiet book ", Author: "An Author", Kind: "book", Status: "planned", Total: 12, Collection: "Favorites"})
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "A quiet book" || item.Revision != 1 {
		t.Fatalf("unexpected item: %+v", item)
	}
	item.Current = 3
	item.Status = "reading"
	updated, err := s.Update(ctx, item)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Update(ctx, item); !errors.Is(err, library.ErrConflict) {
		t.Fatalf("stale update accepted: %v", err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got, err := s.Get(ctx, item.ID)
	if err != nil || got.Current != 3 || got.Revision != updated.Revision {
		t.Fatalf("persistence: %+v %v", got, err)
	}
	stats, err := s.Stats(ctx)
	if err != nil || stats.Total != 1 || stats.Reading != 1 || len(stats.Collections) != 1 {
		t.Fatalf("stats: %+v %v", stats, err)
	}
	events, err := s.History(ctx)
	if err != nil || len(events) != 1 || events[0].Current != 3 {
		t.Fatalf("history: %+v %v", events, err)
	}
	if err = s.Delete(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	events, err = s.History(ctx)
	if err != nil || len(events) != 0 {
		t.Fatalf("orphan history: %+v %v", events, err)
	}
}
func TestFilteringPaginationAndValidation(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	for _, i := range []library.Item{{Title: "Bravo", Author: "Ada", Kind: "book", Status: "planned"}, {Title: "Alpha", Kind: "manga", Status: "reading"}, {Title: "100%", Kind: "book", Status: "completed"}} {
		if _, err := s.Create(ctx, i); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.List(ctx, library.Filter{Query: "%"})
	if err != nil || page.Total != 1 || page.Items[0].Title != "100%" {
		t.Fatalf("literal search: %+v %v", page, err)
	}
	page, err = s.List(ctx, library.Filter{Kind: "book", Sort: "title", Limit: 1, Offset: 1})
	if err != nil || page.Total != 2 || len(page.Items) != 1 || page.Items[0].Title != "Bravo" {
		t.Fatalf("pagination: %+v %v", page, err)
	}
	for _, i := range []library.Item{{Title: "", Kind: "book", Status: "planned"}, {Title: "Title", Kind: "invalid", Status: "planned"}, {Title: "Title", Kind: "book", Status: "planned", Current: 3, Total: 2}} {
		if _, err := s.Create(ctx, i); !errors.Is(err, library.ErrInvalid) {
			t.Fatalf("invalid input accepted: %+v %v", i, err)
		}
	}
}
func TestConcurrentUpdatesHaveSingleWinner(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	i, err := s.Create(ctx, library.Item{Title: "Title", Kind: "book", Status: "planned", Total: 10})
	if err != nil {
		t.Fatal(err)
	}
	i.Current = 1
	results := make(chan error, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { _, err := s.Update(ctx, i); results <- err })
	}
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, library.ErrConflict) {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatalf("got %d successful concurrent updates", winners)
	}
}

func TestLegacyAndFutureDatabasesAreRejected(t *testing.T) {
	for _, statement := range []string{`CREATE TABLE documents (id INTEGER PRIMARY KEY)`, `PRAGMA user_version=2`} {
		path := filepath.Join(t.TempDir(), "library.db")
		s, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
		if err = s.Close(); err != nil {
			t.Fatal(err)
		}
		if reopened, err := Open(path); err == nil {
			_ = reopened.Close()
			t.Fatal("incompatible database accepted")
		}
	}
}
