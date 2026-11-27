// Package seed gives testdb.NewSeeded its data: the imported history plus the
// dev seed of SPEC §11.9, loaded by migratehistory.Load (plan M1-D6).
//
// A package whose tests need that real data is an external _test package
// whose TestMain is seed.Main, and each test takes its own copy with
// testdb.NewSeeded:
//
//	package report_test
//
//	import (
//		"testing"
//
//		"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
//		"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
//	)
//
//	func TestMain(m *testing.M) { seed.Main(m) }
//
//	func TestSomething(t *testing.T) {
//		d := testdb.NewSeeded(t) // a fresh copy of il_seed; d.App is the il_app pool
//		_ = d
//	}
//
// Every other package keeps func TestMain(m *testing.M) { testdb.Main(m) } and
// testdb.New(t), and never pays for the seed.
//
// The package lives apart from testdb, and its users must be external test
// packages, because migratehistory imports auth and calc, whose in-package
// tests import testdb: testdb (or an in-package test of auth or calc)
// importing migratehistory would be an import cycle.
package seed
