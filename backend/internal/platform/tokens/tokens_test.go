package tokens

import (
	"bytes"
	"crypto/hkdf"
	"crypto/sha256"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

var (
	secret = bytes.Repeat([]byte{7}, 32)
	now    = time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
)

func mustKey(t *testing.T, p Purpose) Key {
	t.Helper()
	k, err := DeriveKey(secret, p)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestSignVerify(t *testing.T) {
	undo := mustKey(t, PurposeUndo)
	fields := []string{"7c9e6679-7425-40de-944b-e07fc1f90ae7", "user-1"}
	tok, err := Sign(undo, fields, now.Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("valid", func(t *testing.T) {
		got, err := Verify(undo, tok, now)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, fields) {
			t.Fatalf("fields = %q", got)
		}
	})
	t.Run("valid one second before expiry", func(t *testing.T) {
		if _, err := Verify(undo, tok, now.Add(10*time.Minute-time.Second)); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("expired at exp", func(t *testing.T) {
		if _, err := Verify(undo, tok, now.Add(10*time.Minute)); !errors.Is(err, ErrExpired) {
			t.Fatalf("err = %v, want ErrExpired", err)
		}
	})
	t.Run("tampered payload", func(t *testing.T) {
		p64, s64, _ := strings.Cut(tok, ".")
		payload, _ := b64.DecodeString(p64)
		forged := strings.Replace(string(payload), "user-1", "user-2", 1)
		bad := b64.EncodeToString([]byte(forged)) + "." + s64
		if _, err := Verify(undo, bad, now); !errors.Is(err, ErrBadSignature) {
			t.Fatalf("err = %v, want ErrBadSignature", err)
		}
	})
	t.Run("tampered expiry", func(t *testing.T) {
		p64, s64, _ := strings.Cut(tok, ".")
		payload, _ := b64.DecodeString(p64)
		i := strings.LastIndex(string(payload), sep)
		forged := string(payload[:i+1]) + "9999999999"
		if _, err := Verify(undo, b64.EncodeToString([]byte(forged))+"."+s64, now); !errors.Is(err, ErrBadSignature) {
			t.Fatalf("err = %v, want ErrBadSignature", err)
		}
	})
	t.Run("tampered signature", func(t *testing.T) {
		bad := tok[:len(tok)-2] + "AA"
		if bad == tok {
			bad = tok[:len(tok)-2] + "BB"
		}
		if _, err := Verify(undo, bad, now); !errors.Is(err, ErrBadSignature) {
			t.Fatalf("err = %v, want ErrBadSignature", err)
		}
	})
	t.Run("other purpose rejects it", func(t *testing.T) {
		for _, p := range []Purpose{PurposeVoiceToken, PurposeVoicePending} {
			if _, err := Verify(mustKey(t, p), tok, now); !errors.Is(err, ErrBadSignature) {
				t.Fatalf("%s: err = %v, want ErrBadSignature", p, err)
			}
		}
	})
	t.Run("other secret rejects it", func(t *testing.T) {
		k, _ := DeriveKey(bytes.Repeat([]byte{8}, 32), PurposeUndo)
		if _, err := Verify(k, tok, now); !errors.Is(err, ErrBadSignature) {
			t.Fatalf("err = %v, want ErrBadSignature", err)
		}
	})
	t.Run("malformed", func(t *testing.T) {
		for _, bad := range []string{"", "abc", "a.b.c", "!!!.!!!", b64.EncodeToString([]byte("x")) + ".@@"} {
			if _, err := Verify(undo, bad, now); err == nil {
				t.Fatalf("%q verified", bad)
			}
		}
	})
}

func TestNoFields(t *testing.T) {
	k := mustKey(t, PurposeVoicePending)
	tok, err := Sign(k, nil, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	got, err := Verify(k, tok, now)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestSignRejects(t *testing.T) {
	k := mustKey(t, PurposeUndo)
	if _, err := Sign(k, []string{"a|b"}, now); err == nil {
		t.Fatal("field with | signed")
	}
	if _, err := Sign(Key{}, nil, now); err == nil {
		t.Fatal("zero key signed")
	}
	if _, err := Verify(Key{}, "a.b", now); err == nil {
		t.Fatal("zero key verified")
	}
}

func TestDeriveKey(t *testing.T) {
	if _, err := DeriveKey(secret[:31], PurposeUndo); err == nil {
		t.Fatal("31-byte secret accepted")
	}
	if _, err := DeriveKey(secret, ""); err == nil {
		t.Fatal("empty purpose accepted")
	}
	k := mustKey(t, PurposeVoiceToken)
	want, _ := hkdf.Key(sha256.New, secret, nil, "il voice-token v1", 32)
	if !bytes.Equal(k.k, want) {
		t.Fatal("key is not HKDF-SHA256(secret, info = purpose label)")
	}
	if bytes.Equal(mustKey(t, PurposeUndo).k, k.k) {
		t.Fatal("two purposes share a key")
	}
	if strings.Contains(k.String(), string(k.k)) || k.Purpose() != PurposeVoiceToken {
		t.Fatalf("String/Purpose = %s / %s", k, k.Purpose())
	}
}
