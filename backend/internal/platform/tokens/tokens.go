// Package tokens signs and verifies the short-lived tokens of the API
// (SPEC §8.5): voiceToken, the clarify pending token and the undo tokens.
// Each kind has its own HMAC key derived from APP_SECRET with HKDF-SHA256 and
// a purpose label, so a token of one kind never verifies as another.
//
// Token = base64url(payload) "." base64url(HMAC-SHA256(key, payload)), where
// payload = field "|" field "|" … "|" exp (unix seconds). The voice helpers
// (M12) build their fields as §8.5 lists them.
package tokens

import (
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Purpose is the HKDF info label of one token kind (SPEC §8.5).
type Purpose string

// The purpose labels of SPEC §8.5. Changing one invalidates every token of
// that kind.
const (
	PurposeVoiceToken   Purpose = "il voice-token v1"
	PurposeVoicePending Purpose = "il voice-pending v1"
	PurposeUndo         Purpose = "il undo v1"
)

// MinSecretLen is the least number of bytes APP_SECRET must have.
const MinSecretLen = 32

const sep = "|"

var b64 = base64.RawURLEncoding.Strict()

// Errors returned by Verify. Callers treat them all as "not a valid token";
// they are distinct for tests and logs.
var (
	ErrMalformed    = errors.New("tokens: malformed token")
	ErrBadSignature = errors.New("tokens: bad signature")
	ErrExpired      = errors.New("tokens: expired")
)

// Key is an HMAC key for one purpose.
type Key struct {
	purpose Purpose
	k       []byte
}

// DeriveKey derives the key of purpose from the application secret with
// HKDF-SHA256 (no salt, info = the purpose label, 32 bytes).
func DeriveKey(secret []byte, purpose Purpose) (Key, error) {
	if len(secret) < MinSecretLen {
		return Key{}, fmt.Errorf("tokens: secret has %d bytes, need at least %d", len(secret), MinSecretLen)
	}
	if purpose == "" {
		return Key{}, errors.New("tokens: empty purpose")
	}
	k, err := hkdf.Key(sha256.New, secret, nil, string(purpose), sha256.Size)
	if err != nil {
		return Key{}, fmt.Errorf("tokens: derive key: %w", err)
	}
	return Key{purpose: purpose, k: k}, nil
}

// Purpose returns the purpose the key was derived for.
func (k Key) Purpose() Purpose { return k.purpose }

// String never shows the key material.
func (k Key) String() string { return "tokens.Key(" + string(k.purpose) + ")" }

// Sign returns a token carrying fields that expires at exp. A field must not
// contain "|".
func Sign(k Key, fields []string, exp time.Time) (string, error) {
	if len(k.k) == 0 {
		return "", errors.New("tokens: zero Key")
	}
	for _, f := range fields {
		if strings.Contains(f, sep) {
			return "", fmt.Errorf("tokens: field contains %q", sep)
		}
	}
	payload := strings.Join(append(append([]string(nil), fields...), strconv.FormatInt(exp.Unix(), 10)), sep)
	return b64.EncodeToString([]byte(payload)) + "." + b64.EncodeToString(mac(k, []byte(payload))), nil
}

// Verify checks the token's signature with k (constant-time) and that it has
// not expired at now (valid while now < exp), and returns its fields.
func Verify(k Key, token string, now time.Time) ([]string, error) {
	if len(k.k) == 0 {
		return nil, errors.New("tokens: zero Key")
	}
	p64, s64, ok := strings.Cut(token, ".")
	if !ok {
		return nil, ErrMalformed
	}
	payload, err := b64.DecodeString(p64)
	if err != nil {
		return nil, ErrMalformed
	}
	sig, err := b64.DecodeString(s64)
	if err != nil {
		return nil, ErrMalformed
	}
	if !hmac.Equal(sig, mac(k, payload)) {
		return nil, ErrBadSignature
	}
	parts := strings.Split(string(payload), sep)
	exp, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return nil, ErrMalformed
	}
	if !now.Before(time.Unix(exp, 0)) {
		return nil, ErrExpired
	}
	return parts[:len(parts)-1], nil
}

func mac(k Key, payload []byte) []byte {
	h := hmac.New(sha256.New, k.k)
	h.Write(payload)
	return h.Sum(nil)
}
