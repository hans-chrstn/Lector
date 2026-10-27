package extensions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	root       string
	state      string
	executable string
	mu         sync.RWMutex
	refresh    sync.Mutex
	plugins    map[string]snapshot
	grants     map[string]grant
	slots      chan struct{}
}

func New(root, state, executable string) (*Manager, error) {
	for _, p := range []string{root, filepath.Dir(state)} {
		if err := os.MkdirAll(p, 0700); err != nil {
			return nil, err
		}
	}
	m := &Manager{root: root, state: state, executable: executable, plugins: map[string]snapshot{}, grants: map[string]grant{}, slots: make(chan struct{}, 2)}
	b, err := os.ReadFile(state)
	if err == nil {
		if err = json.Unmarshal(b, &m.grants); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if m.grants == nil {
		m.grants = map[string]grant{}
	}
	if err = m.Refresh(); err != nil {
		return nil, err
	}
	return m, nil
}
func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func readLocal(root, name string, limit int64) ([]byte, error) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	info, err := r.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("plugin files must be regular files")
	}
	f, err := r.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("plugin file too large")
	}
	return b, nil
}
func (m *Manager) Refresh() error {
	m.refresh.Lock()
	defer m.refresh.Unlock()
	entries, err := os.ReadDir(m.root)
	if err != nil {
		return err
	}
	next := map[string]snapshot{}
	for _, entry := range entries {
		if !entry.IsDir() || !identifier.MatchString(entry.Name()) {
			continue
		}
		id := entry.Name()
		s := snapshot{Info: Info{Manifest: Manifest{ID: id, Name: id, Hosts: []string{}, Actions: []Action{}}}}
		b, err := readLocal(m.root, filepath.Join(id, "manifest.json"), 64<<10)
		if err == nil {
			err = json.Unmarshal(b, &s.Manifest)
		}
		if err == nil && s.ID != id {
			err = errors.New("manifest ID must match directory")
		}
		if err == nil {
			err = s.validate()
		}
		if err == nil {
			var src []byte
			src, err = readLocal(m.root, filepath.Join(id, "init.lua"), 256<<10)
			s.source = string(src)
			s.Generation = digest(append(b, src...))
		}
		if err != nil {
			s.Error = err.Error()
			s.ID = id
			s.Name = id
		}
		hosts := slices.Clone(s.Hosts)
		slices.Sort(hosts)
		permissions, _ := json.Marshal(struct {
			Hosts []string
			UI    bool
		}{hosts, s.UI})
		s.permissions = digest(permissions)
		next[id] = s
	}
	m.mu.Lock()
	m.plugins = next
	m.mu.Unlock()
	return nil
}
func (m *Manager) Watch(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = m.Refresh()
		}
	}
}
func (m *Manager) List() []Info {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []Info{}
	for id, s := range m.plugins {
		g := m.grants[id]
		i := s.Info
		i.Approved = g.Permissions == s.permissions
		i.Enabled = g.Enabled && i.Approved && i.Error == ""
		out = append(out, i)
	}
	slices.SortFunc(out, func(a, b Info) int { return strings.Compare(a.ID, b.ID) })
	return out
}
func (m *Manager) SetEnabled(id, generation string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.plugins[id]
	if !ok {
		return ErrMissing
	}
	if enabled && (s.Error != "" || generation != s.Generation) {
		return errors.New("plugin changed; reload its permissions")
	}
	next := map[string]grant{}
	for k, v := range m.grants {
		next[k] = v
	}
	next[id] = grant{Enabled: enabled, Permissions: s.permissions}
	b, err := json.Marshal(next)
	if err != nil {
		return err
	}
	if err = os.WriteFile(m.state+".tmp", b, 0600); err != nil {
		return err
	}
	if err = os.Rename(m.state+".tmp", m.state); err != nil {
		return err
	}
	m.grants = next
	return nil
}
func (m *Manager) Invoke(ctx context.Context, id, action, input string) (Result, error) {
	if len(input) > 16000 {
		return Result{}, errors.New("input exceeds 16000 bytes")
	}
	m.mu.RLock()
	s, ok := m.plugins[id]
	g := m.grants[id]
	m.mu.RUnlock()
	if !ok {
		return Result{}, ErrMissing
	}
	if !g.Enabled || g.Permissions != s.permissions || s.Error != "" {
		return Result{}, ErrDenied
	}
	if !slices.ContainsFunc(s.Actions, func(a Action) bool { return a.ID == action }) {
		return Result{}, errors.New("unknown action")
	}
	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
	default:
		return Result{}, ErrBusy
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req := Request{Source: s.source, Action: action, Input: input, Hosts: s.Hosts, UI: s.UI}
	b, err := json.Marshal(req)
	if err != nil {
		return Result{}, err
	}
	cmd := exec.CommandContext(ctx, m.executable, "plugin-worker")
	cmd.Env = []string{"GOMEMLIMIT=128MiB", "GOMAXPROCS=1"}
	cmd.Stdin = bytes.NewReader(b)
	out := &limitedBuffer{limit: 2 << 20}
	cmd.Stdout = out
	cmd.Stderr = &limitedBuffer{limit: 4096}
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return Result{}, errors.New("plugin exceeded its time limit")
		}
		return Result{}, errors.New("plugin failed; check its code and permissions")
	}
	var result Result
	if err = json.Unmarshal(out.Bytes(), &result); err != nil {
		return result, errors.New("invalid plugin response")
	}
	return result, result.validate(s.UI)
}

type limitedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > b.limit {
		return 0, errors.New("output limit exceeded")
	}
	return b.Buffer.Write(p)
}
