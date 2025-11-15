package redis

import "time"

// RealClock is a clock that uses the real time.
type RealClock struct{}

// Now returns the current time.
func (c *RealClock) Now() time.Time {
	return time.Now()
}
