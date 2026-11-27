package httpx

import (
	"net/http"
	"strings"
)

// MuxFallback serves mux, except that a path at or under prefix ("/api")
// that no pattern serves gets the §7.0a envelope instead of the mux's plain
// text: 404 NOT_FOUND, or 405 METHOD_NOT_ALLOWED with the mux's Allow header
// when the path has routes for other methods. Every other response, the
// mux's redirects (path cleaning, trailing slash) and paths outside prefix
// included, is the mux's own.
//
// It registers no pattern, so the route list (and the §3.3 matrix test)
// doesn't see it. ServeMux.Handler returns the empty pattern for its 404, 405
// and some redirects alike, so the unmatched handler is run against a
// throw-away writer to learn which one it is (NOTES-verify "ServeMux 404 vs
// 405"); that handler only sets headers and writes a short text.
func MuxFallback(mux *http.ServeMux, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != prefix && !strings.HasPrefix(r.URL.Path, prefix+"/") {
			mux.ServeHTTP(w, r)
			return
		}
		h, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r) // sets r.Pattern for the access log
			return
		}
		probe := &probeWriter{header: http.Header{}}
		h.ServeHTTP(probe, r)
		switch probe.status {
		case http.StatusNotFound:
			WriteError(w, NotFound())
		case http.StatusMethodNotAllowed:
			w.Header().Set("Allow", probe.header.Get("Allow"))
			WriteError(w, MethodNotAllowed())
		default:
			mux.ServeHTTP(w, r)
		}
	})
}

// probeWriter records the status and headers of the mux's unmatched handler
// and drops its body.
type probeWriter struct {
	header http.Header
	status int
}

func (p *probeWriter) Header() http.Header { return p.header }

func (p *probeWriter) WriteHeader(code int) {
	if p.status == 0 {
		p.status = code
	}
}

func (p *probeWriter) Write(b []byte) (int, error) {
	if p.status == 0 {
		p.status = http.StatusOK
	}
	return len(b), nil
}
