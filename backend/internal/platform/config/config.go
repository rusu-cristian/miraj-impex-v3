// Package config reads the environment once at start-up (SPEC §11.10).
// Every missing required variable and every invalid value is reported in one
// error; secrets never appear in errors, logs or fmt output.
package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/tokens"
)

// Command is the subcommand the configuration is loaded for; it decides
// which variables are required.
type Command string

// The commands that read the environment.
const (
	Serve       Command = "serve"        // api serve
	Migrate     Command = "migrate"      // api migrate up|status
	CreateOwner Command = "create-owner" // cmd/create-owner (DATABASE_URL, Q6)
)

// required lists the variables each command can't start without.
var required = map[Command][]string{
	Serve:       {"DATABASE_URL", "MAINT_DATABASE_URL", "APP_ORIGIN", "APP_SECRET"},
	Migrate:     {"MIGRATE_DATABASE_URL"},
	CreateOwner: {"DATABASE_URL"},
}

// Env is APP_ENV.
type Env string

// The APP_ENV values.
const (
	Production Env = "production"
	Staging    Env = "staging"
	Dev        Env = "dev"
)

// Log formats (LOG_FORMAT).
const (
	FormatJSON = "json"
	FormatText = "text"
)

// Defaults.
const (
	DefaultDBMaxConns = 10
	DefaultListenAddr = ":8080"
	maxDBMaxConns     = 1000
)

// LookupFunc is os.LookupEnv, or a map lookup in tests.
type LookupFunc func(key string) (string, bool)

// Secret is key material. It prints and logs as "[redacted]".
type Secret []byte

const redacted = "[redacted]"

func (s Secret) String() string { return redacted }

// GoString keeps %#v from printing the bytes.
func (s Secret) GoString() string { return redacted }

// LogValue keeps slog from printing the bytes.
func (s Secret) LogValue() slog.Value { return slog.StringValue(redacted) }

// Config is the process configuration. Thresholds and limits are not here:
// they live in app_setting (SPEC §4.11).
type Config struct {
	Env                Env
	DatabaseURL        string // il_app pool (serve, create-owner)
	MaintDatabaseURL   string // il_maint 1-connection pool (serve)
	MigrateDatabaseURL string // owner role (migrate only)
	DBMaxConns         int32
	ListenAddr         string
	AppOrigin          string // scheme://host[:port], lower-case, no path
	CORSEnabled        bool
	TrustedProxies     []netip.Prefix
	AppSecret          Secret // ≥ 32 bytes, HKDF root for tokens (§8.5)
	LogLevel           slog.Level
	LogFormat          string // json | text
}

// Error lists every missing required variable and every invalid value. It
// names variables only, never their values.
type Error struct {
	Missing []string
	Invalid []string // "NAME: reason"
}

func (e *Error) Error() string {
	var parts []string
	if len(e.Missing) > 0 {
		parts = append(parts, "missing required environment variables: "+strings.Join(e.Missing, ", "))
	}
	if len(e.Invalid) > 0 {
		parts = append(parts, "invalid environment variables: "+strings.Join(e.Invalid, "; "))
	}
	return "config: " + strings.Join(parts, "; ")
}

// Load reads the configuration for cmd through lookup. A value that is set
// but blank counts as missing.
func Load(lookup LookupFunc, cmd Command) (Config, error) {
	need, ok := required[cmd]
	if !ok {
		return Config{}, fmt.Errorf("config: unknown command %q", cmd)
	}
	get := func(key string) string {
		v, _ := lookup(key)
		return strings.TrimSpace(v)
	}
	e := &Error{}
	for _, key := range need {
		if get(key) == "" {
			e.Missing = append(e.Missing, key)
		}
	}
	invalid := func(key, reason string) { e.Invalid = append(e.Invalid, key+": "+reason) }

	// Each DSN is read only by the commands that use it: serve never holds the
	// owner DSN, migrate never holds the app's.
	dsn := func(key string) string {
		if slices.Contains(need, key) {
			return get(key)
		}
		return ""
	}
	c := Config{
		Env:                Production,
		DatabaseURL:        dsn("DATABASE_URL"),
		MaintDatabaseURL:   dsn("MAINT_DATABASE_URL"),
		MigrateDatabaseURL: dsn("MIGRATE_DATABASE_URL"),
		DBMaxConns:         DefaultDBMaxConns,
		ListenAddr:         DefaultListenAddr,
		LogLevel:           slog.LevelInfo,
	}

	if v := get("APP_ENV"); v != "" {
		switch Env(v) {
		case Production, Staging, Dev:
			c.Env = Env(v)
		default:
			invalid("APP_ENV", "must be production, staging or dev")
		}
	}

	if v := get("DB_MAX_CONNS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxDBMaxConns {
			invalid("DB_MAX_CONNS", fmt.Sprintf("must be a whole number from 1 to %d", maxDBMaxConns))
		} else {
			c.DBMaxConns = int32(n)
		}
	}

	if v := get("LISTEN_ADDR"); v != "" {
		if _, _, err := net.SplitHostPort(v); err != nil {
			invalid("LISTEN_ADDR", "must be host:port or :port")
		} else {
			c.ListenAddr = v
		}
	}

	if v := get("APP_ORIGIN"); v != "" {
		origin, err := parseOrigin(v)
		if err != nil {
			invalid("APP_ORIGIN", err.Error())
		} else {
			c.AppOrigin = origin
		}
	}

	if v := get("CORS_ENABLED"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			invalid("CORS_ENABLED", "must be true or false")
		} else {
			c.CORSEnabled = b
		}
	}

	if v := get("TRUSTED_PROXIES"); v != "" {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			p, err := netip.ParsePrefix(part)
			if err != nil {
				invalid("TRUSTED_PROXIES", fmt.Sprintf("%q is not a CIDR (e.g. 10.0.0.0/8)", part))
				continue
			}
			c.TrustedProxies = append(c.TrustedProxies, p.Masked())
		}
	}

	if v := get("APP_SECRET"); v != "" {
		secret, ok := decodeSecret(v)
		switch {
		case !ok:
			invalid("APP_SECRET", fmt.Sprintf("must be base64 of at least %d random bytes", tokens.MinSecretLen))
		case (c.Env == Production || c.Env == Staging) && bytes.Equal(secret, exampleSecret):
			invalid("APP_SECRET", "is the dev placeholder from .env.example; generate one with: openssl rand -base64 32")
		default:
			c.AppSecret = secret
		}
	}

	if v := get("LOG_LEVEL"); v != "" {
		if err := c.LogLevel.UnmarshalText([]byte(v)); err != nil {
			invalid("LOG_LEVEL", "must be debug, info, warn or error")
		}
	}

	c.LogFormat = FormatJSON
	if c.Env == Dev {
		c.LogFormat = FormatText
	}
	if v := get("LOG_FORMAT"); v != "" {
		switch v {
		case FormatJSON, FormatText:
			c.LogFormat = v
		default:
			invalid("LOG_FORMAT", "must be json or text")
		}
	}

	if len(e.Missing) > 0 || len(e.Invalid) > 0 {
		return Config{}, e
	}
	return c, nil
}

// exampleSecret is the APP_SECRET of backend/.env.example (base64 of these
// bytes). It is public, so production and staging refuse it.
var exampleSecret = []byte("dev-only-app-secret-do-not-use-in-prod!!")

// parseOrigin accepts "scheme://host[:port]" with scheme http or https and
// nothing else, and returns it lower-cased and without a default port
// (browsers send it that way).
func parseOrigin(v string) (string, error) {
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" ||
		u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return "", fmt.Errorf("must be an origin like https://app.example (scheme and host, no path or trailing slash)")
	}
	host := u.Host
	if p := u.Port(); (u.Scheme == "https" && p == "443") || (u.Scheme == "http" && p == "80") {
		host = strings.TrimSuffix(host, ":"+p) // browsers omit the default port in Origin
	}
	return strings.ToLower(u.Scheme + "://" + host), nil
}

// decodeSecret accepts standard or URL-safe base64, padded or not.
func decodeSecret(v string) (Secret, bool) {
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(v); err == nil {
			return Secret(b), len(b) >= tokens.MinSecretLen
		}
	}
	return nil, false
}

// LogValue is the configuration with every secret redacted: DSN passwords
// and APP_SECRET.
func (c Config) LogValue() slog.Value {
	proxies := make([]string, len(c.TrustedProxies))
	for i, p := range c.TrustedProxies {
		proxies[i] = p.String()
	}
	secret := ""
	if len(c.AppSecret) > 0 {
		secret = redacted
	}
	return slog.GroupValue(
		slog.String("app_env", string(c.Env)),
		slog.String("database_url", RedactDSN(c.DatabaseURL)),
		slog.String("maint_database_url", RedactDSN(c.MaintDatabaseURL)),
		slog.String("migrate_database_url", RedactDSN(c.MigrateDatabaseURL)),
		slog.Int("db_max_conns", int(c.DBMaxConns)),
		slog.String("listen_addr", c.ListenAddr),
		slog.String("app_origin", c.AppOrigin),
		slog.Bool("cors_enabled", c.CORSEnabled),
		slog.String("trusted_proxies", strings.Join(proxies, ",")),
		slog.String("app_secret", secret),
		slog.String("log_level", c.LogLevel.String()),
		slog.String("log_format", c.LogFormat),
	)
}

// String is LogValue as "key=value" pairs, so fmt never prints a secret.
func (c Config) String() string {
	var b strings.Builder
	for i, a := range c.LogValue().Group() {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(a.Key + "=" + a.Value.String())
	}
	return b.String()
}

// GoString keeps %#v from printing secrets.
func (c Config) GoString() string { return "config.Config{" + c.String() + "}" }

var kvPassword = regexp.MustCompile(`(?i)(password\s*=\s*)('(?:[^'\\]|\\.)*'|\S+)`)

// RedactDSN hides the password of a PostgreSQL DSN in URL form
// (postgres://user:pw@host/db?password=pw) or keyword/value form
// (host=db password=pw). A URL it can't parse is redacted entirely.
func RedactDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	if strings.Contains(dsn, "://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return redacted
		}
		if q := u.Query(); q.Has("password") {
			q.Set("password", "xxxxx")
			u.RawQuery = q.Encode()
		}
		return u.Redacted()
	}
	return kvPassword.ReplaceAllString(dsn, "${1}xxxxx")
}
