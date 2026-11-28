package prices_test

import (
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// TestMain serves the whole test binary: the in-package tests take
// testdb.New, the seeded ones (external package) testdb.NewSeeded. It lives
// here because seed imports farms and golden, so only an external test
// package may import it.
func TestMain(m *testing.M) { seed.Main(m) }
