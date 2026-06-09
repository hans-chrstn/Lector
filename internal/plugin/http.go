package plugin

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/user/lector/internal/core/httpclient"
)

type NetworkProfile struct {
	Name      string
	UserAgent string
	Headers   map[string]string
}

var (
	Profiles = map[string]NetworkProfile{
		"standard": {
			Name:      "Standard Desktop",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Headers: map[string]string{
				"Accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
				"Accept-Language":    "en-US,en;q=0.9",
				"Sec-CH-UA":          `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
				"Sec-CH-UA-Mobile":   "?0",
				"Sec-CH-UA-Platform": `"Windows"`,
			},
		},
		"mobile": {
			Name:      "Standard Mobile",
			UserAgent: "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
			Headers: map[string]string{
				"Accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
				"Accept-Language":    "en-US,en;q=0.9",
				"Sec-CH-UA":          `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
				"Sec-CH-UA-Mobile":   "?1",
				"Sec-CH-UA-Platform": `"Android"`,
			},
		},
	}
)

type ClientManager struct {
	clients map[string]*http.Client
	mu      sync.Mutex
}

var GlobalClientManager = &ClientManager{
	clients: make(map[string]*http.Client),
}

type CadenceManager struct {
	lastRequest map[string]time.Time
	mu          sync.Mutex
}

var GlobalCadenceManager = &CadenceManager{
	lastRequest: make(map[string]time.Time),
}

func (m *CadenceManager) Wait(pluginName string, minMs, maxMs int) {
	m.mu.Lock()
	last := m.lastRequest[pluginName]
	m.mu.Unlock()

	elapsed := time.Since(last)
	rangeMs := maxMs - minMs
	if rangeMs < 1 {
		rangeMs = 1
	}
	delay := time.Duration(minMs+int(time.Now().UnixNano()%int64(rangeMs))) * time.Millisecond

	if elapsed < delay {
		time.Sleep(delay - elapsed)
	}

	m.mu.Lock()
	m.lastRequest[pluginName] = time.Now()
	m.mu.Unlock()
}

func (m *ClientManager) GetClient(pluginName string, profile NetworkProfile) *http.Client {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[pluginName]; ok {
		return client
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: httpclient.GlobalTransport,
		Jar:       jar,
		Timeout:   30 * time.Second,
	}

	m.clients[pluginName] = client
	return client
}

func (s *LuaPlugin) Fetch(method, u, postData, referer string, isAjax bool, customHeaders map[string]string) string {
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	profileName := s.NetworkProfileName
	if profileName == "" {
		profileName = "standard"
	}
	profile := Profiles[profileName]
	if profile.Name == "" {
		profile = Profiles["standard"]
	}

	GlobalCadenceManager.Wait(name, 200, 600)

	if !s.HasCapability("network") {
		fmt.Printf("[Security] [%s] Blocked unauthorized network request (Capability 'network' not enabled)\n", name)
		return "ERROR: Network access not enabled"
	}

	parsed, err := url.Parse(u)
	if err != nil {
		fmt.Printf("[Security] [%s] Invalid URL: %s\n", name, u)
		return "ERROR: Invalid URL"
	}

	allowed := false
	if len(s.Permissions) == 0 {
		allowed = true
	} else {
		for _, domain := range s.Permissions {
			if domain == "*" || strings.HasSuffix(parsed.Host, domain) {
				allowed = true
				break
			}
		}
	}

	if !allowed {
		fmt.Printf("[Security] [%s] Blocked unauthorized fetch to %s (Domain not in permissions)\n", name, parsed.Host)
		return "ERROR: Unauthorized domain"
	}

	client := GlobalClientManager.GetClient(name, profile)

	var body io.Reader
	if method == "POST" {
		body = strings.NewReader(postData)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", profile.UserAgent)
	for k, v := range profile.Headers {
		req.Header.Set(k, v)
	}

	if isAjax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	} else {
		req.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Network] [%s] Request failed to %s: %v\n", name, u, err)
		return ""
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(resp.Body)
	content := string(out)

	if resp.StatusCode == 403 {
		fmt.Printf("[Network] [%s] Compatibility warning (403) at %s\n", name, u)
	}

	return content
}

func (s *LuaPlugin) Download(u, destPath, referer string, customHeaders map[string]string) bool {
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	profileName := s.NetworkProfileName
	if profileName == "" {
		profileName = "standard"
	}
	profile := Profiles[profileName]
	if profile.Name == "" {
		profile = Profiles["standard"]
	}

	GlobalCadenceManager.Wait(name, 200, 600)

	if !s.HasCapability("network") {
		return false
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}

	allowed := false
	if len(s.Permissions) == 0 {
		allowed = true
	} else {
		for _, domain := range s.Permissions {
			if domain == "*" || strings.HasSuffix(parsed.Host, domain) {
				allowed = true
				break
			}
		}
	}

	if !allowed {
		return false
	}

	client := GlobalClientManager.GetClient(name, profile)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", profile.UserAgent)
	for k, v := range profile.Headers {
		req.Header.Set(k, v)
	}

	if referer != "" {
		req.Header.Set("Referer", referer)
	} else {
		req.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")
	}

	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return false
	}

	baseDir, _ := filepath.Abs("downloads")
	baseDir = filepath.Clean(baseDir)
	fullPath, _ := filepath.Abs(destPath)
	fullPath = filepath.Clean(fullPath)

	if fullPath != baseDir && !strings.HasPrefix(fullPath, baseDir+string(filepath.Separator)) {
		return false
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return false
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err == nil
}
