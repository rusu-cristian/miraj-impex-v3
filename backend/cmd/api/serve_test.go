package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestServeMissingEnv(t *testing.T) {
	t.Parallel()
	none := func(string) (string, bool) { return "", false }
	code, out, errOut := runCmd(t, none, "serve")
	if code != exitError {
		t.Fatalf("exit %d", code)
	}
	// One line naming every missing variable (SPEC §11.10).
	if out != "" || strings.Count(errOut, "\n") != 1 {
		t.Fatalf("want one line on stderr, got stdout %q stderr %q", out, errOut)
	}
	for _, v := range []string{"DATABASE_URL", "MAINT_DATABASE_URL", "APP_ORIGIN", "APP_SECRET"} {
		if !strings.Contains(errOut, v) {
			t.Fatalf("%s not named in %q", v, errOut)
		}
	}
	if strings.Contains(errOut, "MIGRATE_DATABASE_URL") {
		t.Fatalf("serve must not need MIGRATE_DATABASE_URL: %q", errOut)
	}

	code, _, errOut = runCmd(t, none, "serve", "now")
	if code != exitUsage || !strings.Contains(errOut, "usage: api serve") {
		t.Fatalf("extra argument: exit %d, %q", code, errOut)
	}
}

// serveEnv is a complete serve environment on database name (superuser DSN:
// runServe opens its own pools), listening on addr.
func serveEnv(name, addr string) func(string) (string, bool) {
	env := map[string]string{
		"APP_ENV":            "dev",
		"DATABASE_URL":       testdb.DSN(name),
		"MAINT_DATABASE_URL": testdb.DSN(name),
		"APP_ORIGIN":         testOrigin,
		"APP_SECRET":         "ZGV2LW9ubHktYXBwLXNlY3JldC1kby1ub3QtdXNlLWluLXByb2QhIQ==", // padded base64, as .env.example
		"LISTEN_ADDR":        addr,
		"LOG_FORMAT":         "json",
	}
	return func(k string) (string, bool) { v, ok := env[k]; return v, ok }
}

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// fetch GETs url without keep-alive and returns the status and body.
func fetch(url string) (int, string, error) {
	c := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := c.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b), err
}

// TestRunServe runs the real `api serve` (config, pools, wiring, listener)
// and stops it by cancelling its context, as SIGTERM does.
func TestRunServe(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, errOut := &logBuffer{}, &logBuffer{}
	done := make(chan int, 1)
	go func() { done <- run(ctx, []string{"serve"}, serveEnv(d.Name, addr), out, errOut) }()

	waitFor(t, "/readyz 200", func() bool {
		code, body, err := fetch("http://" + addr + "/readyz")
		return err == nil && code == http.StatusOK && body == "ok"
	})
	if code, body, err := fetch("http://" + addr + "/healthz"); err != nil || code != http.StatusOK || body != "ok" {
		t.Fatalf("healthz: %d %q %v", code, body, err)
	}
	cancel()
	select {
	case code := <-done:
		if code != exitOK {
			t.Fatalf("exit %d\n%s%s", code, out, errOut)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("serve did not stop")
	}
	logs := out.String()
	for _, want := range []string{`"msg":"api serve listening"`, `"msg":"shutting down"`, `"msg":"api serve stopped"`, `"app_secret":"[redacted]"`} {
		if !strings.Contains(logs, want) {
			t.Fatalf("log lacks %s:\n%s", want, logs)
		}
	}
	if strings.Contains(logs, "ZGV2LW9ubHkt") {
		t.Fatal("APP_SECRET logged")
	}
}

func TestRunServeListenFails(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	code, out, _ := runCmd(t, serveEnv(d.Name, busy.Addr().String()), "serve")
	if code != exitError || !strings.Contains(out, `"msg":"serve: listen"`) {
		t.Fatalf("exit %d, %s", code, out)
	}
}

// slowApp is newTestApp with one more route, /test/slow, that blocks until
// release is closed and then pings the app pool.
type slowApp struct {
	*app
	entered chan struct{}
	release chan struct{}
	pingErr chan error // the in-flight request's ping
}

func newSlowApp(t *testing.T) (*slowApp, *logBuffer) {
	t.Helper()
	a, _, logs := newTestApp(t)
	s := &slowApp{app: a, entered: make(chan struct{}), release: make(chan struct{}), pingErr: make(chan error, 1)}
	inner := a.handler
	a.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test/slow" {
			inner.ServeHTTP(w, r)
			return
		}
		close(s.entered)
		<-s.release
		s.pingErr <- a.pools.App.Ping(context.Background())
		_, _ = io.WriteString(w, "done")
	})
	return s, logs
}

type fetched struct {
	code int
	body string
	err  error
}

func TestServeGracefulShutdown(t *testing.T) {
	t.Parallel()
	s, logs := newSlowApp(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	served := make(chan error, 1)
	go func() { served <- s.serve(ctx, ln) }()

	if code, body, err := fetch("http://" + addr + "/readyz"); err != nil || code != http.StatusOK || body != "ok" {
		t.Fatalf("readyz before shutdown: %d %q %v", code, body, err)
	}
	slow := make(chan fetched, 1)
	go func() {
		code, body, err := fetch("http://" + addr + "/test/slow")
		slow <- fetched{code, body, err}
	}()
	<-s.entered

	cancel() // SIGTERM
	// Shutdown has started once the listener refuses new connections ...
	waitFor(t, "the listener to close", func() bool {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
		}
		return err != nil
	})
	// ... and by then /readyz already answers 503.
	wantText(t, get(t, s.handler, "/readyz"), http.StatusServiceUnavailable, "shutdown")
	select {
	case err := <-served:
		t.Fatalf("serve returned (%v) before the in-flight request finished", err)
	default:
	}

	close(s.release)
	if r := <-slow; r.err != nil || r.code != http.StatusOK || r.body != "done" {
		t.Fatalf("in-flight request: %+v", r)
	}
	if err := <-s.pingErr; err != nil {
		t.Fatalf("pool closed before the in-flight request finished: %v", err)
	}
	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("serve: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not return")
	}
	// The pools close last.
	if err := s.pools.App.Ping(context.Background()); err == nil {
		t.Fatal("app pool still open after serve returned")
	}
	if err := s.pools.Maint.Ping(context.Background()); err == nil {
		t.Fatal("maintenance pool still open after serve returned")
	}
	if !strings.Contains(logs.String(), `"msg":"shutting down","budget":"20s"`) {
		t.Fatalf("no shutdown line with the 20 s budget:\n%s", logs.String())
	}
}

func TestServeShutdownBudget(t *testing.T) {
	t.Parallel()
	s, logs := newSlowApp(t)
	s.shutdownTimeout = 200 * time.Millisecond
	defer close(s.release)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- s.serve(ctx, ln) }()
	slow := make(chan fetched, 1)
	go func() {
		code, body, err := fetch("http://" + ln.Addr().String() + "/test/slow")
		slow <- fetched{code, body, err}
	}()
	<-s.entered

	cancel()
	select {
	case err := <-served:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("serve: %v, want the budget's deadline", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve ignored the shutdown budget")
	}
	// The stuck request's connection was closed.
	if r := <-slow; r.err == nil {
		t.Fatalf("stuck request finished: %+v", r)
	}
	if err := s.pools.App.Ping(context.Background()); err == nil {
		t.Fatal("app pool still open after serve returned")
	}
	if !strings.Contains(logs.String(), "budget ran out") {
		t.Fatalf("no error line:\n%s", logs.String())
	}
}
