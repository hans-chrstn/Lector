package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

var (
	GlobalTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				host = addr
				port = ""
			}

			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}

			var targetIP net.IP
			for _, ip := range ips {
				if isPrivateIP(ip) {
					return nil, fmt.Errorf("security: access to private IP %s is blocked", ip.String())
				}
				if targetIP == nil {
					targetIP = ip
				}
			}

			if targetIP == nil {
				return nil, fmt.Errorf("security: no valid IP found for host %s", host)
			}

			dialAddr := targetIP.String()
			if port != "" {
				dialAddr = net.JoinHostPort(targetIP.String(), port)
			}

			dialer := &net.Dialer{
				Timeout:   15 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return dialer.DialContext(ctx, network, dialAddr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
		},
	}
	InternalClient *http.Client
	RelaxedClient  *http.Client
)

func init() {
	InternalClient = &http.Client{
		Transport: GlobalTransport,
		Timeout:   30 * time.Second,
	}
	RelaxedClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
				NextProtos:         []string{"h2", "http/1.1"},
			},
		},
		Timeout: 30 * time.Second,
	}
}

var BypassPrivateIPCheckForTesting = false

func isPrivateIP(ip net.IP) bool {
	if BypassPrivateIPCheckForTesting {
		return false
	}
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
	return false
}
