package redis

import "time"

// MockClock is a mock clock that can be used in tests.
type MockClock struct {
	now time.Time
}

// Now returns the current time.
func (c *MockClock) Now() time.Time {
	return c.now
}

// SetNow sets the current time.
func (c *MockClock) SetNow(t time.Time) {
	c.now = t
}
