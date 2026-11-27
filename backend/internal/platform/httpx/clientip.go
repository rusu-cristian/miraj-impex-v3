package httpx

import (
	"context"
	"net/http"
	"net/netip"
	"strings"
)

// client is what ClientIP learns about the caller.
type client struct {
	ip    netip.Addr
	https bool // the client's connection is HTTPS (TLS here, or a trusted X-Forwarded-Proto)
}

// ClientIP finds the caller's address (SPEC §11.4): r.RemoteAddr, or, when the
// request comes from a trusted proxy, the right-most address of
// X-Forwarded-For that is not itself a trusted proxy. X-Forwarded-For and
// X-Forwarded-Proto from any other peer are ignored, so a client can't spoof
// its address or scheme. Read the results with ClientIPFrom and IsHTTPS.
func ClientIP(trusted []netip.Prefix) Middleware {
	isTrusted := func(a netip.Addr) bool {
		for _, p := range trusted {
			if p.Contains(a) {
				return true
			}
		}
		return false
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := client{ip: parseAddr(r.RemoteAddr), https: r.TLS != nil}
			if c.ip.IsValid() && isTrusted(c.ip) {
				c.ip = forwardedFor(r.Header, c.ip, isTrusted)
				if proto, ok := lastValue(r.Header, "X-Forwarded-Proto"); ok {
					c.https = strings.EqualFold(proto, "https")
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), clientKey, c)))
		})
	}
}

// forwardedFor walks X-Forwarded-For from the right, starting at the trusted
// peer: each trusted hop vouches for the address to its left. It stops at the
// first untrusted address (the client) or at an entry it can't parse (keeping
// the last good hop).
func forwardedFor(h http.Header, peer netip.Addr, isTrusted func(netip.Addr) bool) netip.Addr {
	var hops []string
	for _, v := range h.Values("X-Forwarded-For") {
		hops = append(hops, strings.Split(v, ",")...)
	}
	addr := peer
	for i := len(hops) - 1; i >= 0; i-- {
		a := parseAddr(strings.TrimSpace(hops[i]))
		if !a.IsValid() {
			break
		}
		addr = a
		if !isTrusted(a) {
			break
		}
	}
	return addr
}

// lastValue is the right-most comma-separated value of a header: the one the
// nearest (trusted) proxy wrote.
func lastValue(h http.Header, key string) (string, bool) {
	vs := h.Values(key)
	if len(vs) == 0 {
		return "", false
	}
	parts := strings.Split(vs[len(vs)-1], ",")
	v := strings.TrimSpace(parts[len(parts)-1])
	return v, v != ""
}

// parseAddr reads "ip", "ip:port" or "[ipv6]:port"; IPv4-mapped IPv6
// addresses become IPv4. It returns the zero Addr when s is none of these.
func parseAddr(s string) netip.Addr {
	if ap, err := netip.ParseAddrPort(s); err == nil {
		return ap.Addr().WithZone("").Unmap()
	}
	if a, err := netip.ParseAddr(s); err == nil {
		return a.WithZone("").Unmap()
	}
	return netip.Addr{}
}

// ClientIPFrom is the caller's address found by ClientIP; the zero Addr when
// ClientIP didn't run.
func ClientIPFrom(ctx context.Context) netip.Addr {
	c, _ := ctx.Value(clientKey).(client)
	return c.ip
}

// IsHTTPS reports whether the client reached us over HTTPS: TLS on this
// connection, or X-Forwarded-Proto: https from a trusted proxy (ClientIP).
func IsHTTPS(r *http.Request) bool {
	if c, ok := r.Context().Value(clientKey).(client); ok {
		return c.https
	}
	return r.TLS != nil
}
