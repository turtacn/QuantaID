package redis

import "time"

// Clock is an interface for getting the current time.
type Clock interface {
	Now() time.Time
}
