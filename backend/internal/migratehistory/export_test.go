package migratehistory

// WithCheck returns opts with an extra Verify check: the test hook that makes
// an import fail its verification (M1 T5).
func WithCheck(opts Options, c Check) Options {
	opts.extraChecks = append(opts.extraChecks, c)
	return opts
}

// ExtractsDir is the extracts directory the tests read.
var ExtractsDir = extractsDir
