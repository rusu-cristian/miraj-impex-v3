package httpx

// Validatable is a request body with level-1 shape checks (SPEC §7.0b).
type Validatable interface {
	Validate(v *Validator)
}

// Validator collects shape issues. It never stops at the first one: issues
// keep the order they were added in, so a Validate method that checks fields
// in document order reports them in document order and the client can mark
// every field at once.
type Validator struct {
	issues []Issue
}

// Add records an issue.
func (v *Validator) Add(path, code, message string) {
	v.issues = append(v.issues, Issue{Path: path, Code: code, Message: message})
}

// Check records an issue when ok is false and reports ok.
func (v *Validator) Check(ok bool, path, code, message string) bool {
	if !ok {
		v.Add(path, code, message)
	}
	return ok
}

// Valid reports whether no issue was recorded.
func (v *Validator) Valid() bool { return len(v.issues) == 0 }

// Issues returns the recorded issues in order (a copy).
func (v *Validator) Issues() []Issue { return append([]Issue(nil), v.issues...) }

// Err is nil when no issue was recorded, else 400 VALIDATION carrying every
// issue, with the first issue's message as error.message.
func (v *Validator) Err() error {
	if v.Valid() {
		return nil
	}
	return Validation(v.Issues()...)
}

// Validate runs x.Validate on a fresh Validator and returns its Err.
func Validate(x Validatable) error {
	var v Validator
	x.Validate(&v)
	return v.Err()
}
