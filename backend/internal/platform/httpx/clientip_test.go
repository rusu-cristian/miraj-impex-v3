package httpx

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestClientIP(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("127.0.0.1/32")}
	tests := []struct {
		name   string
		remote string
		xff    []string
		want   string
	}{
		{"direct, no header", "203.0.113.7:51000", nil, "203.0.113.7"},
		{"spoofed XFF from an untrusted peer is ignored", "203.0.113.7:51000", []string{"1.2.3.4"}, "203.0.113.7"},
		{"trusted proxy", "10.0.0.2:4000", []string{"198.51.100.9"}, "198.51.100.9"},
		{"client spoofs a left entry; right-most untrusted wins", "10.0.0.2:4000", []string{"1.2.3.4, 198.51.100.9"}, "198.51.100.9"},
		{"two trusted hops", "127.0.0.1:4000", []string{"1.2.3.4, 198.51.100.9, 10.1.2.3"}, "198.51.100.9"},
		{"several header lines", "10.0.0.2:4000", []string{"1.2.3.4", "198.51.100.9"}, "198.51.100.9"},
		{"every hop trusted: left-most", "10.0.0.2:4000", []string{"10.9.9.9, 10.1.1.1"}, "10.9.9.9"},
		{"unparsable entry stops the walk", "10.0.0.2:4000", []string{"198.51.100.9, junk, 10.1.1.1"}, "10.1.1.1"},
		{"entry with a port", "10.0.0.2:4000", []string{"198.51.100.9:6000"}, "198.51.100.9"},
		{"IPv6 client", "10.0.0.2:4000", []string{"2001:db8::1"}, "2001:db8::1"},
		{"trusted proxy without XFF", "10.0.0.2:4000", nil, "10.0.0.2"},
		{"IPv4-mapped peer", "[::ffff:203.0.113.7]:51000", nil, "203.0.113.7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got netip.Addr
			h := ClientIP(trusted)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = ClientIPFrom(r.Context())
			}))
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remote
			for _, v := range tt.xff {
				r.Header.Add("X-Forwarded-For", v)
			}
			h.ServeHTTP(httptest.NewRecorder(), r)
			if got.String() != tt.want {
				t.Errorf("client IP %s, want %s", got, tt.want)
			}
		})
	}
}

func TestClientIPWithoutTrustedProxies(t *testing.T) {
	var got netip.Addr
	h := ClientIP(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = ClientIPFrom(r.Context()) }))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.2:4000"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	h.ServeHTTP(httptest.NewRecorder(), r)
	if got.String() != "10.0.0.2" {
		t.Errorf("got %s", got)
	}
	if ClientIPFrom(r.Context()).IsValid() {
		t.Error("ClientIPFrom without the middleware should be the zero Addr")
	}
}

func TestIsHTTPS(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	tests := []struct {
		name, remote, proto string
		tls                 bool
		want                bool
	}{
		{"trusted proxy says https", "10.0.0.2:1", "https", false, true},
		{"trusted proxy says http", "10.0.0.2:1", "http", false, false},
		{"proxy appends to a spoofed value", "10.0.0.2:1", "http, https", false, true},
		{"untrusted peer can't claim https", "203.0.113.7:1", "https", false, false},
		{"direct TLS", "203.0.113.7:1", "", true, true},
		{"plain direct", "203.0.113.7:1", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			h := ClientIP(trusted)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = IsHTTPS(r) }))
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remote
			if tt.proto != "" {
				r.Header.Set("X-Forwarded-Proto", tt.proto)
			}
			if tt.tls {
				r.TLS = &tls.ConnectionState{}
			}
			h.ServeHTTP(httptest.NewRecorder(), r)
			if got != tt.want {
				t.Errorf("IsHTTPS %v, want %v", got, tt.want)
			}
		})
	}
}
