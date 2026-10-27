package library

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid item")
var ErrNotFound = errors.New("item not found")
var ErrConflict = errors.New("item changed; reload and try again")

type Item struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Kind       string    `json:"kind"`
	Status     string    `json:"status"`
	Current    int       `json:"current"`
	Total      int       `json:"total"`
	Notes      string    `json:"notes"`
	Collection string    `json:"collection"`
	File       string    `json:"file,omitempty"`
	Format     string    `json:"format,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Revision   int       `json:"revision"`
}

func (i *Item) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	i.Author = strings.TrimSpace(i.Author)
	i.Collection = strings.TrimSpace(i.Collection)
	if i.Title == "" || len(i.Title) > 300 || len(i.Author) > 200 || len(i.Collection) > 80 || len(i.Notes) > 20000 {
		return ErrInvalid
	}
	switch i.Kind {
	case "book", "manga", "anime", "video":
	default:
		return ErrInvalid
	}
	switch i.Status {
	case "planned", "reading", "completed", "paused":
	default:
		return ErrInvalid
	}
	if i.Total < 0 || i.Total > 10000000 || i.Current < 0 || (i.Total > 0 && i.Current > i.Total) {
		return ErrInvalid
	}
	return nil
}

type Filter struct {
	Query, Kind, Status, Collection, Sort string
	Limit, Offset                         int
}
type Page struct {
	Items []Item `json:"items"`
	Total int    `json:"total"`
}
type Stats struct {
	Total       int      `json:"total"`
	Reading     int      `json:"reading"`
	Completed   int      `json:"completed"`
	Planned     int      `json:"planned"`
	Collections []string `json:"collections"`
}
type Event struct {
	ID      int64     `json:"id"`
	ItemID  string    `json:"item_id"`
	Title   string    `json:"title"`
	Current int       `json:"current"`
	Status  string    `json:"status"`
	At      time.Time `json:"at"`
}

type Repository interface {
	List(context.Context, Filter) (Page, error)
	Get(context.Context, string) (Item, error)
	Create(context.Context, Item) (Item, error)
	Update(context.Context, Item) (Item, error)
	Delete(context.Context, string) error
	Stats(context.Context) (Stats, error)
	History(context.Context) ([]Event, error)
	Close() error
}
