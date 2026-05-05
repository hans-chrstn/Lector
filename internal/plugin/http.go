package plugin

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	GlobalTransport = &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	HTTPClient = &http.Client{
		Transport: GlobalTransport,
		Timeout:   30 * time.Second,
	}
	HTTPMu sync.Mutex
)

func (s *LuaPlugin) Fetch(method, u, postData, referer string, isAjax bool) string {
	var body io.Reader
	if method == "POST" {
		body = strings.NewReader(postData)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	if isAjax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	HTTPMu.Lock()
	resp, err := HTTPClient.Do(req)
	HTTPMu.Unlock()

	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(resp.Body)
	content := string(out)

	if strings.Contains(content, "cf-browser-verification") || resp.StatusCode == 403 {
		fmt.Printf("[Plugin] Anti-Bot detected at %s\n", u)
	}

	return content
}
