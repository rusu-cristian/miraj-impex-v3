package httpx

import (
	"mime"
	"net/http"
	"slices"
	"strings"
)

// CSRF enforces the Origin rule of SPEC §11.4 on every method other than
// GET, HEAD and OPTIONS:
//   - when Origin is present it must equal appOrigin exactly;
//   - when Origin is absent, Sec-Fetch-Site must be "same-origin";
//   - when both are absent the request is refused.
//
// Refusals are 403 FORBIDDEN. CSRF is global (cmd/api's middleware); the
// Content-Type rule is per route (ContentType), because the guide preview
// takes multipart and voice STT an audio type. http.CrossOriginProtection is
// not used: it lets requests with neither header through, accepts
// Sec-Fetch-Site "none", and compares Origin with the Host header instead of
// APP_ORIGIN (NOTES-verify).
func CSRF(appOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) || allowedOrigin(r, appOrigin) {
				next.ServeHTTP(w, r)
				return
			}
			WriteError(w, Forbidden())
		})
	}
}

// MediaJSON is the media type of every JSON route (SPEC §11.4, §7.0a).
const MediaJSON = "application/json"

// ContentType is the Content-Type rule of SPEC §11.4, per route: on every
// method other than GET, HEAD and OPTIONS the request's media type
// (parameters such as charset ignored, case-insensitive) must be one of
// allowed, else 403 FORBIDDEN. A browser can send text/plain, form and
// multipart bodies cross-site without a preflight; it can't send JSON. JSON
// routes use ContentType(MediaJSON) (cmd/api's public and authed chains); the
// guide preview (multipart) and voice STT (audio) build their own chains.
func ContentType(allowed ...string) Middleware {
	types := make([]string, len(allowed))
	for i, t := range allowed {
		types[i] = strings.ToLower(t)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || !slices.Contains(types, mt) {
				WriteError(w, Forbidden())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isSafeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

func allowedOrigin(r *http.Request, appOrigin string) bool {
	origins := r.Header.Values("Origin")
	switch len(origins) {
	case 0:
		return r.Header.Get("Sec-Fetch-Site") == "same-origin"
	case 1:
		return appOrigin != "" && origins[0] == appOrigin
	default:
		return false
	}
}

// The CORS values of SPEC §1 (cross-origin deployment).
const (
	corsAllowMethods = "GET, POST, PUT, PATCH, DELETE"
	corsAllowHeaders = "Content-Type, Last-Event-ID, X-Voice-Run"
	corsMaxAge       = "600"
)

// CORS answers cross-origin requests from appOrigin only, when enabled
// (CORS_ENABLED, SPEC §1). A request whose Origin equals appOrigin gets
// Access-Control-Allow-Origin: <appOrigin>, Access-Control-Allow-Credentials:
// true; a preflight from it also gets the allowed methods and headers and
// Access-Control-Max-Age: 600 and is answered 204 here. Every other origin
// gets no CORS headers, and its preflight is refused with 403 FORBIDDEN.
// Every response carries Vary: Origin while CORS is enabled (it depends on
// the Origin). When disabled, CORS does nothing.
func CORS(appOrigin string, enabled bool) Middleware {
	return func(next http.Handler) http.Handler {
		if !enabled {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Add("Vary", "Origin")
			origins := r.Header.Values("Origin")
			allowed := len(origins) == 1 && appOrigin != "" && origins[0] == appOrigin
			preflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
			if allowed {
				h.Set("Access-Control-Allow-Origin", appOrigin)
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			switch {
			case preflight && allowed:
				h.Set("Access-Control-Allow-Methods", corsAllowMethods)
				h.Set("Access-Control-Allow-Headers", corsAllowHeaders)
				h.Set("Access-Control-Max-Age", corsMaxAge)
				w.WriteHeader(http.StatusNoContent)
			case preflight:
				WriteError(w, Forbidden())
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
