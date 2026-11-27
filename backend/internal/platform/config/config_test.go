package config

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"slices"
	"strings"
	"testing"
)

const (
	pw      = "s3cr3t-pw"
	rawKey  = "0123456789abcdef0123456789abcdef!" // 33 bytes
	shortPw = "x"
)

var secretB64 = base64.StdEncoding.EncodeToString([]byte(rawKey))

func env(m map[string]string) LookupFunc {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func serveEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL":       "postgres://il_app:" + pw + "@db:5432/impex?sslmode=require",
		"MAINT_DATABASE_URL": "host=db user=il_maint password=" + pw + " dbname=impex",
		"APP_ORIGIN":         "https://App.Example",
		"APP_SECRET":         secretB64,
	}
}

func TestLoadServeDefaults(t *testing.T) {
	c, err := Load(env(serveEnv()), Serve)
	if err != nil {
		t.Fatal(err)
	}
	if c.DBMaxConns != 10 || c.ListenAddr != ":8080" {
		t.Errorf("defaults: DBMaxConns=%d ListenAddr=%q", c.DBMaxConns, c.ListenAddr)
	}
	if c.Env != Production || c.LogFormat != FormatJSON || c.LogLevel != slog.LevelInfo || c.CORSEnabled {
		t.Errorf("defaults: %v", c)
	}
	if c.AppOrigin != "https://app.example" {
		t.Errorf("AppOrigin = %q", c.AppOrigin)
	}
	if string(c.AppSecret) != rawKey {
		t.Error("AppSecret not decoded")
	}
	if len(c.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies = %v", c.TrustedProxies)
	}
}

func TestLoadAll(t *testing.T) {
	m := serveEnv()
	m["APP_ENV"] = "dev"
	m["DB_MAX_CONNS"] = "4"
	m["LISTEN_ADDR"] = "127.0.0.1:9090"
	m["CORS_ENABLED"] = "true"
	m["TRUSTED_PROXIES"] = "10.0.0.0/8, 127.0.0.1/32,,fd00::/8"
	m["LOG_LEVEL"] = "debug"
	c, err := Load(env(m), Serve)
	if err != nil {
		t.Fatal(err)
	}
	want := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("127.0.0.1/32"), netip.MustParsePrefix("fd00::/8")}
	if !slices.Equal(c.TrustedProxies, want) {
		t.Errorf("TrustedProxies = %v", c.TrustedProxies)
	}
	if c.Env != Dev || c.LogFormat != FormatText || c.LogLevel != slog.LevelDebug || c.DBMaxConns != 4 ||
		c.ListenAddr != "127.0.0.1:9090" || !c.CORSEnabled {
		t.Errorf("got %v", c)
	}
	m["LOG_FORMAT"] = "json"
	if c, _ := Load(env(m), Serve); c.LogFormat != FormatJSON {
		t.Errorf("LOG_FORMAT override = %q", c.LogFormat)
	}
}

func TestLoadMissingAllAtOnce(t *testing.T) {
	for _, tc := range []struct {
		cmd  Command
		env  map[string]string
		want []string
	}{
		{Serve, map[string]string{}, []string{"DATABASE_URL", "MAINT_DATABASE_URL", "APP_ORIGIN", "APP_SECRET"}},
		{Serve, map[string]string{"DATABASE_URL": "postgres://x", "APP_SECRET": "  "}, []string{"MAINT_DATABASE_URL", "APP_ORIGIN", "APP_SECRET"}},
		{Migrate, map[string]string{"DATABASE_URL": "postgres://x"}, []string{"MIGRATE_DATABASE_URL"}},
		{CreateOwner, map[string]string{}, []string{"DATABASE_URL"}},
	} {
		t.Run(string(tc.cmd), func(t *testing.T) {
			_, err := Load(env(tc.env), tc.cmd)
			var ce *Error
			if !errors.As(err, &ce) {
				t.Fatalf("err = %v", err)
			}
			if !slices.Equal(ce.Missing, tc.want) {
				t.Fatalf("Missing = %v, want %v", ce.Missing, tc.want)
			}
			for _, name := range tc.want {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("message %q does not name %s", err, name)
				}
			}
		})
	}
	// migrate needs nothing that serve needs.
	if _, err := Load(env(map[string]string{"MIGRATE_DATABASE_URL": "postgres://o:p@db/impex"}), Migrate); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := Load(env(nil), "bogus"); err == nil {
		t.Fatal("unknown command accepted")
	}
}

func TestLoadInvalid(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"APP_SECRET", base64.StdEncoding.EncodeToString([]byte("only-31-bytes-long-secret-value"))},
		{"APP_SECRET", "not base64 !!"},
		{"TRUSTED_PROXIES", "10.0.0.0/8,10.0.0.0/33"},
		{"TRUSTED_PROXIES", "10.0.0.1"},
		{"DB_MAX_CONNS", "0"},
		{"DB_MAX_CONNS", "ten"},
		{"LISTEN_ADDR", "8080"},
		{"APP_ORIGIN", "https://app.example/"},
		{"APP_ORIGIN", "app.example"},
		{"APP_ORIGIN", "ftp://app.example"},
		{"APP_ENV", "prod"},
		{"CORS_ENABLED", "yes please"},
		{"LOG_LEVEL", "loud"},
		{"LOG_FORMAT", "xml"},
	} {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			m := serveEnv()
			m[tc.key] = tc.value
			_, err := Load(env(m), Serve)
			var ce *Error
			if !errors.As(err, &ce) || len(ce.Invalid) != 1 || !strings.HasPrefix(ce.Invalid[0], tc.key+":") {
				t.Fatalf("err = %v", err)
			}
		})
	}
	// Missing and invalid come back together, and the error never shows a secret.
	m := map[string]string{"APP_SECRET": base64.StdEncoding.EncodeToString([]byte(shortPw)), "DB_MAX_CONNS": "-1"}
	_, err := Load(env(m), Serve)
	var ce *Error
	if !errors.As(err, &ce) || len(ce.Missing) != 3 || len(ce.Invalid) != 2 {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(err.Error(), m["APP_SECRET"]) {
		t.Fatal("error shows the secret")
	}
}

func TestLogValueRedacts(t *testing.T) {
	c, err := Load(env(serveEnv()), Serve)
	if err != nil {
		t.Fatal(err)
	}
	mc, err := Load(env(map[string]string{
		"MIGRATE_DATABASE_URL": "postgresql://il_owner@db/impex?password=" + pw + "&sslmode=require",
	}), Migrate)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("config", "config", c, "migrate", mc)
	slog.New(slog.NewTextHandler(&buf, nil)).Info("config", "config", c, "secret", c.AppSecret, "migrate", mc)
	out := buf.String() + fmt.Sprintf("%v %+v %#v %s %v %#v %v %#v", c, c, c, c.AppSecret, c.AppSecret, c.AppSecret, mc, mc)
	for _, leak := range []string{pw, secretB64, rawKey} {
		if strings.Contains(out, leak) {
			t.Fatalf("output contains %q:\n%s", leak, out)
		}
	}
	for _, want := range []string{"il_app", "db:5432", "il_maint", "il_owner", "https://app.example", `"app_secret":"[redacted]"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestRedactDSN(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"postgres://u:" + pw + "@h:5432/d?sslmode=require", "postgres://u:xxxxx@h:5432/d?sslmode=require"},
		{"postgres://u@h/d", "postgres://u@h/d"},
		{"postgres://u@h/d?password=" + pw, "postgres://u@h/d?password=xxxxx"},
		{"host=h password=" + pw + " user=u", "host=h password=xxxxx user=u"},
		{"host=h password = '" + pw + " with space' user=u", "host=h password = xxxxx user=u"},
		{"postgres://u:pa%zz@h/d", "[redacted]"},
		{"", ""},
	} {
		if got := RedactDSN(tc.in); got != tc.want {
			t.Errorf("RedactDSN(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLoadReadsOnlyTheCommandsDSNs(t *testing.T) {
	all := serveEnv()
	all["MIGRATE_DATABASE_URL"] = "postgres://il_owner:" + pw + "@db/impex"
	tests := []struct {
		cmd                   Command
		app, maint, migration bool
	}{
		{Serve, true, true, false},
		{Migrate, false, false, true},
		{CreateOwner, true, false, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.cmd), func(t *testing.T) {
			c, err := Load(env(all), tt.cmd)
			if err != nil {
				t.Fatal(err)
			}
			if (c.DatabaseURL != "") != tt.app || (c.MaintDatabaseURL != "") != tt.maint || (c.MigrateDatabaseURL != "") != tt.migration {
				t.Errorf("DATABASE_URL %v, MAINT_DATABASE_URL %v, MIGRATE_DATABASE_URL %v; want %v %v %v",
					c.DatabaseURL != "", c.MaintDatabaseURL != "", c.MigrateDatabaseURL != "", tt.app, tt.maint, tt.migration)
			}
		})
	}
}

// exampleEnvSecret is APP_SECRET as backend/.env.example sets it.
func exampleEnvSecret(t *testing.T) string {
	t.Helper()
	f, err := os.Open("../../../.env.example")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if v, ok := strings.CutPrefix(sc.Text(), "APP_SECRET="); ok {
			return v
		}
	}
	t.Fatal(".env.example has no APP_SECRET line")
	return ""
}

func TestLoadRefusesExampleSecretOutsideDev(t *testing.T) {
	example := exampleEnvSecret(t)
	for _, appEnv := range []string{"production", "staging", ""} { // unset = production
		t.Run("APP_ENV="+appEnv, func(t *testing.T) {
			m := serveEnv()
			m["APP_SECRET"] = example
			if appEnv != "" {
				m["APP_ENV"] = appEnv
			}
			_, err := Load(env(m), Serve)
			var ce *Error
			if !errors.As(err, &ce) || len(ce.Invalid) != 1 || !strings.HasPrefix(ce.Invalid[0], "APP_SECRET: ") {
				t.Fatalf("err = %v, want APP_SECRET invalid", err)
			}
			if strings.Contains(err.Error(), example) {
				t.Error("the error prints the secret")
			}
		})
	}
	m := serveEnv()
	m["APP_SECRET"] = example
	m["APP_ENV"] = "dev"
	if _, err := Load(env(m), Serve); err != nil {
		t.Errorf("dev with the example secret: %v", err)
	}
}

func TestOriginDropsDefaultPort(t *testing.T) {
	for in, want := range map[string]string{
		"https://app.example:443":  "https://app.example",
		"http://app.example:80":    "http://app.example",
		"https://App.Example:8443": "https://app.example:8443",
		"http://app.example:443":   "http://app.example:443",
		"https://app.example:80":   "https://app.example:80",
		"https://[::1]:443":        "https://[::1]",
		"https://app.example":      "https://app.example",
	} {
		got, err := parseOrigin(in)
		if err != nil || got != want {
			t.Errorf("parseOrigin(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}
