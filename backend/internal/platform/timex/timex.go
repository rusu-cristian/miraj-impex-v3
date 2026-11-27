// Package timex holds the business time zone, the injected clock and the
// civil-day and local-instant types of SPEC §11.1 and §7.0a.
//
// This is the only package that may read the wall clock (time.Now and
// friends); everything else takes a Clock.
package timex

import (
	"sync"
	"time"
	_ "time/tzdata" // embed the zone database: the distroless image has no /usr/share/zoneinfo
)

// Chisinau is the business time zone (Europe/Chisinau, EET/EEST). It is
// loaded once when the package initialises; a failure stops the process.
var Chisinau = mustLoad("Europe/Chisinau")

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("timex: load time zone " + name + ": " + err.Error())
	}
	return loc
}

// Clock is the source of "now" for services and handlers.
type Clock interface {
	Now() time.Time
}

// RealClock reads the wall clock.
type RealClock struct{}

// Now returns the current time.
func (RealClock) Now() time.Time { return time.Now() }

// Fixed is a Clock frozen at one instant that tests can move. It is safe for
// concurrent use.
type Fixed struct {
	mu sync.Mutex
	t  time.Time
}

// FixedClock returns a Clock frozen at t.
func FixedClock(t time.Time) *Fixed { return &Fixed{t: t} }

// Now returns the frozen instant.
func (c *Fixed) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

// Set moves the clock to t.
func (c *Fixed) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = t
}

// Advance moves the clock forward by d (backward when d < 0).
func (c *Fixed) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// Today is the civil day of the clock's now in Europe/Chisinau.
func Today(c Clock) Day { return DayOf(c.Now()) }

// DayOf is the civil day of t in Europe/Chisinau.
func DayOf(t time.Time) Day {
	y, m, d := t.In(Chisinau).Date()
	return Day{y: y, m: m, d: d}
}
