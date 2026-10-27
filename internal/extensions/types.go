package extensions

import (
	"errors"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

var identifier = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
var ErrDenied = errors.New("plugin is disabled or permissions need approval")
var ErrMissing = errors.New("plugin not found")
var ErrBusy = errors.New("all plugin workers are busy; try again")

type Manifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Hosts       []string `json:"hosts"`
	UI          bool     `json:"ui"`
	Actions     []Action `json:"actions"`
}
type Action struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type Panel struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Tone  string `json:"tone"`
}
type Entry struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Kind   string `json:"kind"`
	Total  int    `json:"total"`
	Text   string `json:"text,omitempty"`
}
type Result struct {
	Panels  []Panel `json:"panels"`
	Entries []Entry `json:"entries"`
}
type Info struct {
	Manifest
	Enabled    bool   `json:"enabled"`
	Approved   bool   `json:"approved"`
	Error      string `json:"error,omitempty"`
	Generation string `json:"generation"`
}
type grant struct {
	Enabled     bool   `json:"enabled"`
	Permissions string `json:"permissions"`
}
type snapshot struct {
	Info
	source      string
	permissions string
}
type Request struct {
	Source string   `json:"source"`
	Action string   `json:"action"`
	Input  string   `json:"input"`
	Hosts  []string `json:"hosts"`
	UI     bool     `json:"ui"`
}

func (m Manifest) validate() error {
	if !identifier.MatchString(m.ID) || m.Name == "" || len(m.Name) > 100 || len(m.Description) > 1000 || len(m.Version) > 40 || len(m.Hosts) > 20 || len(m.Actions) > 20 {
		return errors.New("invalid manifest")
	}
	seen := map[string]bool{}
	for _, a := range m.Actions {
		if !identifier.MatchString(a.ID) || a.Label == "" || len(a.Label) > 80 || seen[a.ID] {
			return errors.New("invalid or duplicate action")
		}
		seen[a.ID] = true
	}
	for _, h := range m.Hosts {
		u, err := url.Parse("https://" + h)
		if err != nil || u.Hostname() != h || u.Port() != "" || u.User != nil || strings.ContainsAny(h, "/*?#:@") || h != strings.ToLower(h) || !strings.Contains(h, ".") {
			return errors.New("hosts must be exact lowercase domain names")
		}
	}
	return nil
}
func (r Result) validate(ui bool) error {
	if len(r.Panels) > 20 || len(r.Entries) > 100 {
		return errors.New("plugin returned too many results")
	}
	if len(r.Panels) > 0 && !ui {
		return ErrDenied
	}
	for _, p := range r.Panels {
		if len(p.Title) > 200 || len(p.Text) > 20000 || !slices.Contains([]string{"", "neutral", "accent"}, p.Tone) {
			return errors.New("invalid panel")
		}
	}
	for _, e := range r.Entries {
		if e.Title == "" || len(e.Title) > 300 || len(e.Author) > 200 || len(e.Text) > 1<<20 || e.Total < 0 || e.Total > 10000000 || !slices.Contains([]string{"book", "manga", "anime", "video"}, e.Kind) {
			return errors.New("invalid entry")
		}
	}
	return nil
}
