package main

import (
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// seed.Main is testdb.Main plus the seeded template, built only when a test
// calls testdb.NewSeeded (the M2 farm reads on real data). package main is
// imported by nothing, so in-package tests here can use seed without the
// import cycle that keeps other packages' seeded tests external.
func TestMain(m *testing.M) { seed.Main(m) }
