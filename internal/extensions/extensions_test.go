package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSandboxAndStructuredResults(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req := Request{UI: true, Action: "run", Source: `function run(input)
 assert(os == nil and io == nil and require == nil and debug == nil and loadfile == nil and loadstring == nil)
 return {panels={{title="Hello",text=input,tone="accent"}}}
 end`, Input: "<script>alert(1)</script>"}
	result, err := execute(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Panels) != 1 || result.Panels[0].Text != req.Input {
		t.Fatalf("unexpected result %+v", result)
	}
	req.UI = false
	if _, err = execute(ctx, req); !errors.Is(err, ErrDenied) {
		t.Fatalf("UI permission bypass: %v", err)
	}
}
func TestRunawayAndCyclicPlugins(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := execute(ctx, Request{Action: "run", Source: `function run() while true do end end`})
	if err == nil {
		t.Fatal("runaway plugin did not stop")
	}
	_, err = execute(context.Background(), Request{Action: "run", Source: `function run() local t={} t.self=t return t end`})
	if err == nil {
		t.Fatal("cyclic result accepted")
	}
}
func TestNetworkPolicy(t *testing.T) {
	for _, addr := range []string{"127.0.0.1", "10.1.2.3", "192.168.1.1", "169.254.169.254", "100.64.0.1", "::1", "::ffff:127.0.0.1", "2001:db8::1", "64:ff9b::a00:1", "3fff::1", "0.0.0.0", "224.0.0.1", "198.18.0.1"} {
		if publicAddress(netip.MustParseAddr(addr)) {
			t.Errorf("private/reserved address allowed: %s", addr)
		}
	}
	if !publicAddress(netip.MustParseAddr("8.8.8.8")) {
		t.Fatal("public address denied")
	}
	for _, raw := range []string{"http://example.org", "https://example.org:444", "https://example.org.evil.test", "https://user:pass@example.org", "file:///etc/passwd"} {
		if _, err := allowedURL(raw, []string{"example.org"}); err == nil {
			t.Errorf("URL allowed: %s", raw)
		}
	}
	if _, err := allowedURL("https://example.org/a", []string{"example.org"}); err != nil {
		t.Fatal(err)
	}
}
func TestHotReloadPermissionChangesAndPersistence(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(filepath.Join(root, "demo"), 0700); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{ID: "demo", Name: "Demo", Version: "1", UI: true, Actions: []Action{{ID: "run", Label: "Run"}}}
	write := func() {
		b, _ := json.Marshal(manifest)
		if err := os.WriteFile(filepath.Join(root, "demo", "manifest.json"), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write()
	if err := os.WriteFile(filepath.Join(root, "demo", "init.lua"), []byte(`function run() return {panels={{title="Hi",text="One"}}} end`), 0600); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(dir, "grants.json")
	m, err := New(root, state, "unused")
	if err != nil {
		t.Fatal(err)
	}
	info := m.List()[0]
	if info.Enabled {
		t.Fatal("new plugin starts enabled")
	}
	if err = m.SetEnabled("demo", info.Generation, true); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "demo", "init.lua"), []byte(`function run() return {panels={{title="Hi",text="Two"}}} end`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.Refresh(); err != nil {
		t.Fatal(err)
	}
	next := m.List()[0]
	if next.Generation == info.Generation || !next.Enabled {
		t.Fatal("code hot reload did not preserve permissions")
	}
	manifest.Hosts = []string{"example.org"}
	write()
	if err = m.Refresh(); err != nil {
		t.Fatal(err)
	}
	if m.List()[0].Enabled || m.List()[0].Approved {
		t.Fatal("permission escalation did not disable plugin")
	}
	if _, err = m.Invoke(context.Background(), "demo", "run", ""); !errors.Is(err, ErrDenied) {
		t.Fatal("unapproved invocation allowed")
	}
	if err = m.SetEnabled("demo", info.Generation, true); err == nil {
		t.Fatal("stale permission review accepted")
	}
	if err = m.SetEnabled("demo", m.List()[0].Generation, true); err != nil {
		t.Fatal(err)
	}
	m, err = New(root, state, "unused")
	if err != nil || !m.List()[0].Enabled {
		t.Fatalf("grant persistence: %v", err)
	}
}
func TestFilesCannotEscapePluginRoot(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "secret")
	if err := os.WriteFile(outside, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "init.lua")); err != nil {
		t.Fatal(err)
	}
	if _, err := readLocal(root, "init.lua", 100); err == nil {
		t.Fatal("symlink allowed")
	}
	if _, err := readLocal(root, "../secret", 100); err == nil {
		t.Fatal("path traversal allowed")
	}
}
func TestIsolatedInvocationAndReload(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("resource-limited workers require Linux")
	}
	dir := t.TempDir()
	root := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(filepath.Join(root, "demo"), 0700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"id":"demo","name":"Demo","ui":true,"actions":[{"id":"run","label":"Run"}]}`
	if err := os.WriteFile(filepath.Join(root, "demo", "manifest.json"), []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	write := func(text string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, "demo", "init.lua"), []byte(`function run(input) return {panels={{title="Result",text="`+text+`"}}} end`), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("first")
	executable := filepath.Join(dir, "lector")
	build := exec.Command("go", "build", "-o", executable, "../../cmd/lector")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build worker: %s %v", output, err)
	}
	m, err := New(root, filepath.Join(dir, "grants.json"), executable)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetEnabled("demo", m.List()[0].Generation, true); err != nil {
		t.Fatal(err)
	}
	result, err := m.Invoke(context.Background(), "demo", "run", "")
	if err != nil || len(result.Panels) != 1 || result.Panels[0].Text != "first" {
		t.Fatalf("worker result: %+v %v", result, err)
	}
	write("second")
	if err = m.Refresh(); err != nil {
		t.Fatal(err)
	}
	result, err = m.Invoke(context.Background(), "demo", "run", "")
	if err != nil || result.Panels[0].Text != "second" {
		t.Fatalf("reload result: %+v %v", result, err)
	}
	if err = os.WriteFile(filepath.Join(root, "demo", "init.lua"), []byte(`function run() local values={} while true do values[#values+1]=string.rep("x",1048576) end end`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.Refresh(); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Invoke(context.Background(), "demo", "run", ""); err == nil {
		t.Fatal("unbounded allocation succeeded")
	}
	write("recovered")
	if err = m.Refresh(); err != nil {
		t.Fatal(err)
	}
	result, err = m.Invoke(context.Background(), "demo", "run", "")
	if err != nil || len(result.Panels) != 1 || result.Panels[0].Text != "recovered" {
		t.Fatalf("server did not recover after worker exhaustion: %+v %v", result, err)
	}
	if err = m.SetEnabled("demo", m.List()[0].Generation, false); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Invoke(context.Background(), "demo", "run", ""); !errors.Is(err, ErrDenied) {
		t.Fatalf("disabled invocation: %v", err)
	}
}
func TestResultLimits(t *testing.T) {
	if err := (Result{Panels: []Panel{{Text: strings.Repeat("x", 20001)}}}).validate(true); err == nil {
		t.Fatal("oversized panel accepted")
	}
	if err := (Result{Entries: []Entry{{Title: "Entry", Kind: "unknown"}}}).validate(false); err == nil {
		t.Fatal("unknown media type accepted")
	}
}
