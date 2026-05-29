package plugin

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
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
)

func (s *LuaPlugin) Fetch(method, u, postData, referer string, isAjax bool) string {
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")

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

	if isPrivateHost(parsed.Host) {
		fmt.Printf("[Security] [%s] Blocked SSRF attempt to private host: %s\n", name, parsed.Host)
		return "ERROR: Unauthorized access to local network"
	}

	var body io.Reader
	if method == "POST" {
		body = strings.NewReader(postData)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,all/all;q=0.8")
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

	resp, err := HTTPClient.Do(req)

	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(resp.Body)
	content := string(out)

	if strings.Contains(content, "cf-browser-verification") || resp.StatusCode == 403 {
		fmt.Printf("[Plugin] [%s] Anti-Bot detected at %s\n", name, u)
	}

	return content
}

func isPrivateHost(host string) bool {
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host
	}

	if hostname == "localhost" {
		return true
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
		if ip4 := ip.To4(); ip4 != nil {
			if ip4[0] == 10 ||
				(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
				(ip4[0] == 192 && ip4[1] == 168) {
				return true
			}
		}
	}
	return false
}
