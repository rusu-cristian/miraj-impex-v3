package auth

import (
	"context"
	"encoding/base64"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/argon2"
)

var phcPattern = regexp.MustCompile(`^\$argon2id\$v=19\$m=19456,t=2,p=1\$([A-Za-z0-9+/]{22})\$([A-Za-z0-9+/]{43})$`)

func TestHashFormatAndVerify(t *testing.T) {
	ctx := context.Background()
	h := newHasher()
	phc, err := h.Hash(ctx, "parola-buna")
	if err != nil {
		t.Fatal(err)
	}
	m := phcPattern.FindStringSubmatch(phc)
	if m == nil {
		t.Fatalf("PHC string %q", phc)
	}
	salt, _ := base64.RawStdEncoding.DecodeString(m[1])
	key, _ := base64.RawStdEncoding.DecodeString(m[2])
	if want := argon2.IDKey([]byte("parola-buna"), salt, 2, 19*1024, 1, 32); string(key) != string(want) {
		t.Error("hash is not argon2id(t=2, m=19456, p=1, 32 bytes) of the password and salt")
	}

	ok, rehash, err := h.Verify(ctx, "parola-buna", phc)
	if err != nil || !ok || rehash {
		t.Errorf("right password: ok=%v rehash=%v err=%v", ok, rehash, err)
	}
	ok, _, err = h.Verify(ctx, "parola-rea", phc)
	if err != nil || ok {
		t.Errorf("wrong password: ok=%v err=%v", ok, err)
	}
	if other, _ := h.Hash(ctx, "parola-buna"); other == phc {
		t.Error("two hashes of one password share a salt")
	}
}

func TestVerifyReadsParamsFromHash(t *testing.T) {
	ctx := context.Background()
	old := Params{Time: 1, Memory: 8 * 1024, Threads: 2, KeyLen: 32}
	salt := []byte("0123456789abcdef")
	phc := encodePHC(old, salt, argon2.IDKey([]byte("veche"), salt, old.Time, old.Memory, old.Threads, old.KeyLen))

	h := newHasher()
	var seen Params
	h.idKey = func(pw, s []byte, tm, mem uint32, th uint8, n uint32) []byte {
		seen = Params{Time: tm, Memory: mem, Threads: th, KeyLen: n}
		return argon2.IDKey(pw, s, tm, mem, th, n)
	}
	ok, rehash, err := h.Verify(ctx, "veche", phc)
	if err != nil || !ok || !rehash {
		t.Errorf("old params: ok=%v rehash=%v err=%v", ok, rehash, err)
	}
	if seen != old {
		t.Errorf("IDKey ran with %+v, want the hash's %+v", seen, old)
	}
	if ok, rehash, _ := h.Verify(ctx, "alta", phc); ok || rehash {
		t.Error("a wrong password must not ask for a rehash")
	}
}

func TestDecodePHCRefusesBadHashes(t *testing.T) {
	good := dummyHash
	for _, bad := range []string{
		"",
		"plain",
		strings.Replace(good, "argon2id", "argon2i", 1),
		strings.Replace(good, "v=19", "v=16", 1),
		strings.Replace(good, "m=19456", "m=99999999", 1), // over the memory cap
		strings.Replace(good, "t=2", "t=0", 1),
		strings.Replace(good, "p=1", "p=0", 1),
		strings.Replace(good, "m=19456,", "x=1,", 1),
		good + "$extra",
		good[:len(good)-4] + "!!!!",
	} {
		if _, _, _, err := decodePHC(bad); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
	if _, _, err := newHasher().Verify(context.Background(), "x", "garbage"); err == nil {
		t.Error("Verify accepted a malformed hash")
	}
}

func TestBurnRunsOneIDKey(t *testing.T) {
	h := newHasher()
	if err := h.burn(context.Background(), "orice"); err != nil {
		t.Fatal(err)
	}
	if n := h.runs.Load(); n != 1 {
		t.Errorf("IDKey runs = %d, want 1", n)
	}
}

func TestHashSemaphoreAllowsFour(t *testing.T) {
	h := newHasher()
	var cur, peak atomic.Int32
	h.idKey = func(_, _ []byte, _, _ uint32, _ uint8, n uint32) []byte {
		c := cur.Add(1)
		for {
			p := peak.Load()
			if c <= p || peak.CompareAndSwap(p, c) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		cur.Add(-1)
		return make([]byte, n)
	}
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			if _, err := h.Hash(context.Background(), "p"); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if p := peak.Load(); p != maxHashers {
		t.Errorf("peak concurrent hashes = %d, want %d", p, maxHashers)
	}
}

func TestHashWaitHonoursContext(t *testing.T) {
	h := newHasher()
	for range maxHashers {
		h.sem <- struct{}{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.Hash(ctx, "p"); err != context.Canceled {
		t.Errorf("err = %v, want context.Canceled while all slots are taken", err)
	}
}

func TestGeneratePassword(t *testing.T) {
	seen := map[string]bool{}
	counts := map[rune]int{}
	for range 300 {
		p, err := GeneratePassword()
		if err != nil {
			t.Fatal(err)
		}
		if len(p) != 12 {
			t.Fatalf("%q has %d characters", p, len(p))
		}
		for _, c := range p {
			if !strings.ContainsRune("abcdefghjkmnpqrstuvwxyz23456789", c) {
				t.Fatalf("%q has %q outside the alphabet", p, c)
			}
			counts[c]++
		}
		if seen[p] {
			t.Fatalf("repeated password %q", p)
		}
		seen[p] = true
	}
	if len(counts) != len(passwordAlphabet) {
		t.Errorf("3600 characters used only %d of %d letters", len(counts), len(passwordAlphabet))
	}
}
