package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/user/lector/internal/core/httpclient"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/repository"
)

func TestLuaEnvironment(t *testing.T) {
	db.InitDB(":memory:")
	luaCode := `
		app.register_manifest({type="utility"})
		app.enable_capability("ui")
		app.add_section("test", "Test")
		assert(type(app) == "table", "app global should be a table")
		assert(type(net) == "table", "net global should be a table")
		assert(type(doc) == "table", "doc global should be a table")
		assert(type(fs) == "table", "fs global should be a table")
		assert(type(http_get) == "function", "http_get global should be a function")
	`
	err := os.WriteFile("test_env.lua", []byte(luaCode), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_env.lua")

	_, err = plugin.NewLuaPlugin("test", "test_env.lua", repository.NewPluginRepository())
	if err != nil {
		t.Fatalf("Lua environment validation failed: %v", err)
	}
}

func TestLuaSandboxSecurity(t *testing.T) {
	db.InitDB(":memory:")
	t.Run("Dangerous Modules Restricted", func(t *testing.T) {
		luaCode := `
			app.register_manifest({type="utility"})
			app.enable_capability("ui")
			app.add_section("test", "Test")
			if os and os.execute then
				error("os.execute is available!")
			end
			if io then
				error("io module is available!")
			end
			if debug then
				error("debug module is available!")
			end
			if package then
				error("package module is available!")
			end
		`
		err := os.WriteFile("test_sandbox.lua", []byte(luaCode), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("test_sandbox.lua")

		_, err = plugin.NewLuaPlugin("test", "test_sandbox.lua", repository.NewPluginRepository())
		if err != nil {
			if strings.Contains(err.Error(), "is available") {
				t.Fatalf("Security failure: dangerous module detected: %v", err)
			}
		}
	})
}

func TestLuaFSSandboxSecurity(t *testing.T) {
	db.InitDB(":memory:")
	defer os.RemoveAll("uploads")

	luaCode := `
		app.register_manifest({type="utility"})
		app.enable_capability("storage")
		app.enable_capability("ui")
		app.add_section("test", "Test")
		local ok1 = fs.write_file("../outside.txt", "should fail")
		local content1 = fs.read_file("../outside.txt")
		local ok2 = fs.write_file("inside.txt", "hello")
		local content2 = fs.read_file("inside.txt")
		assert(not ok1, "writing outside sandbox should return false")
		assert(content1 == nil, "reading outside sandbox should return nil")
		assert(ok2, "writing inside sandbox should succeed")
		assert(content2 == "hello", "reading inside sandbox should match written content")
	`
	err := os.WriteFile("test_fs_sandbox.lua", []byte(luaCode), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_fs_sandbox.lua")

	_, err = plugin.NewLuaPlugin("test", "test_fs_sandbox.lua", repository.NewPluginRepository())
	if err != nil {
		t.Fatalf("FS sandbox security test failed: %v", err)
	}
}

func TestLuaNetworkSandboxSecurity(t *testing.T) {
	db.InitDB(":memory:")
	t.Run("SSRF Dialer Protection", func(t *testing.T) {
		luaCode := `
			app.register_manifest({type="utility"})
			app.enable_capability("network")
			app.enable_capability("ui")
			app.add_section("test", "Test")
			local res = app.net.request("GET", "http://127.0.0.1:8080/")
			assert(res == "", "should block private IP")
		`
		err := os.WriteFile("test_ssrf.lua", []byte(luaCode), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("test_ssrf.lua")

		_, err = plugin.NewLuaPlugin("test", "test_ssrf.lua", repository.NewPluginRepository())
		if err != nil {
			t.Fatalf("SSRF dialer test failed: %v", err)
		}
	})

	t.Run("Unauthorized Domain Protection", func(t *testing.T) {
		luaCode := `
			app.register_manifest({type="utility"})
			app.enable_capability("network")
			app.enable_capability("ui")
			app.add_section("test", "Test")
			app.add_permission("example.com")
			local res = app.net.request("GET", "http://google.com/")
			assert(res == "ERROR: Unauthorized domain", "should block unauthorized domain")
		`
		err := os.WriteFile("test_domain.lua", []byte(luaCode), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("test_domain.lua")

		_, err = plugin.NewLuaPlugin("test", "test_domain.lua", repository.NewPluginRepository())
		if err != nil {
			t.Fatalf("Domain permissions check failed: %v", err)
		}
	})

	t.Run("Custom Headers Support", func(t *testing.T) {
		httpclient.BypassPrivateIPCheckForTesting = true
		defer func() {
			httpclient.BypassPrivateIPCheckForTesting = false
		}()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Custom-Test") == "Value" {
				w.Write([]byte("success"))
			} else {
				w.Write([]byte("failure"))
			}
		}))
		defer ts.Close()

		luaCode := `
			app.register_manifest({type="utility"})
			app.enable_capability("network")
			app.enable_capability("ui")
			app.add_section("test", "Test")
			local res = app.net.request("GET", "` + ts.URL + `", {
				headers = {
					["X-Custom-Test"] = "Value"
				}
			})
			assert(res == "success", "custom headers should be passed: " .. tostring(res))
		`
		err := os.WriteFile("test_headers.lua", []byte(luaCode), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("test_headers.lua")

		_, err = plugin.NewLuaPlugin("test", "test_headers.lua", repository.NewPluginRepository())
		if err != nil {
			t.Fatalf("Custom headers test failed: %v", err)
		}
	})
}
