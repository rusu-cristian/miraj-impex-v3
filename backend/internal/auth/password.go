package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync/atomic"

	"golang.org/x/crypto/argon2"
)

// Params are the argon2id parameters (SPEC §11.4).
type Params struct {
	Time    uint32 // t, passes
	Memory  uint32 // m, KiB
	Threads uint8  // p
	KeyLen  uint32 // hash length in bytes
}

// DefaultParams are the parameters new hashes use: t=2, m=19 MiB, p=1, 32 bytes.
var DefaultParams = Params{Time: 2, Memory: 19 * 1024, Threads: 1, KeyLen: 32}

const (
	saltLen = 16
	// maxHashers bounds concurrent hashes (SPEC §11.4: 4 at a time, 19 MiB each).
	maxHashers = 4
	// Bounds for parameters read back from a stored hash, so a corrupt row
	// can't make one verify allocate gigabytes.
	maxMemory = 256 * 1024
	maxTime   = 16
	maxKeyLen = 64
)

var b64 = base64.RawStdEncoding.Strict()

// dummyHash is a well-formed hash nobody's password matches. A login that
// doesn't exist is verified against it, so it costs one IDKey like a real one
// and timing doesn't reveal which logins exist (SPEC §11.4).
var dummyHash = encodePHC(DefaultParams, make([]byte, saltLen), make([]byte, DefaultParams.KeyLen))

// Hasher hashes and verifies passwords with the §11.4 argon2id parameters, at
// most maxHashers at a time. Every password hash of the backend comes from
// one: the auth service, create-owner and the dev seed.
type Hasher struct {
	params Params
	sem    chan struct{}
	// idKey is argon2.IDKey; tests replace it to watch concurrency.
	idKey func(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte
	// runs counts IDKey calls (a test hook: an unknown login must run one).
	runs atomic.Int64
}

// NewHasher returns a Hasher with DefaultParams. It needs no database.
func NewHasher() *Hasher {
	return &Hasher{params: DefaultParams, sem: make(chan struct{}, maxHashers), idKey: argon2.IDKey}
}

// derive runs one IDKey under the semaphore.
func (h *Hasher) derive(ctx context.Context, password, salt []byte, p Params) ([]byte, error) {
	select {
	case h.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-h.sem }()
	h.runs.Add(1)
	return h.idKey(password, salt, p.Time, p.Memory, p.Threads, p.KeyLen), nil
}

// Hash returns the PHC string of password with the current parameters:
// $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash> (raw standard base64).
func (h *Hasher) Hash(ctx context.Context, password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: salt: %w", err)
	}
	key, err := h.derive(ctx, []byte(password), salt, h.params)
	if err != nil {
		return "", err
	}
	return encodePHC(h.params, salt, key), nil
}

// Verify checks password against a PHC string, with the parameters read from
// the string. rehash is true when the password matches and the string's
// parameters differ from the current ones (re-hash on the next successful
// login, §11.4).
func (h *Hasher) Verify(ctx context.Context, password, phc string) (ok, rehash bool, err error) {
	p, salt, want, err := decodePHC(phc)
	if err != nil {
		return false, false, err
	}
	got, err := h.derive(ctx, []byte(password), salt, p)
	if err != nil {
		return false, false, err
	}
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return false, false, nil
	}
	return true, p != h.params || len(salt) != saltLen, nil
}

// burn runs one IDKey against the dummy hash (unknown login).
func (h *Hasher) burn(ctx context.Context, password string) error {
	_, _, err := h.Verify(ctx, password, dummyHash)
	return err
}

func encodePHC(p Params, salt, key []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Threads, b64.EncodeToString(salt), b64.EncodeToString(key))
}

var errBadHash = errors.New("auth: malformed argon2id hash")

func decodePHC(s string) (p Params, salt, key []byte, err error) {
	parts := strings.Split(s, "$")
	// "", "argon2id", "v=19", "m=…,t=…,p=…", salt, hash
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v="+strconv.Itoa(argon2.Version) {
		return p, nil, nil, errBadHash
	}
	var m, t, threads uint64
	for _, kv := range strings.Split(parts[3], ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return p, nil, nil, errBadHash
		}
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return p, nil, nil, errBadHash
		}
		switch k {
		case "m":
			m = n
		case "t":
			t = n
		case "p":
			threads = n
		default:
			return p, nil, nil, errBadHash
		}
	}
	if salt, err = b64.DecodeString(parts[4]); err != nil || len(salt) < 8 {
		return p, nil, nil, errBadHash
	}
	if key, err = b64.DecodeString(parts[5]); err != nil || len(key) < 16 || len(key) > maxKeyLen {
		return p, nil, nil, errBadHash
	}
	if t < 1 || t > maxTime || threads < 1 || threads > 255 || m < 8*threads || m > maxMemory {
		return p, nil, nil, errBadHash
	}
	return Params{Time: uint32(t), Memory: uint32(m), Threads: uint8(threads), KeyLen: uint32(len(key))}, salt, key, nil
}

// passwordAlphabet has no look-alikes (no i, l, o, 0, 1).
const passwordAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// PasswordLen is the length of a generated password (SPEC §6.16, §11.4).
const PasswordLen = 12

// GeneratePassword returns PasswordLen characters from passwordAlphabet, each
// drawn with crypto/rand.Int (no modulo bias). The result is never logged.
func GeneratePassword() (string, error) {
	var b strings.Builder
	n := big.NewInt(int64(len(passwordAlphabet)))
	for range PasswordLen {
		i, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", fmt.Errorf("auth: generate password: %w", err)
		}
		b.WriteByte(passwordAlphabet[i.Int64()])
	}
	return b.String(), nil
}
