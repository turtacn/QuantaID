package testutils

import (
    "net"

	"github.com/oschwald/geoip2-golang"
)

// MockGeoIPReader is a mock implementation for testing
type MockGeoIPReader struct {
    LookupFunc func(net.IP) (*geoip2.City, error)
}

func (m *MockGeoIPReader) City(ip net.IP) (*geoip2.City, error) {
    if m.LookupFunc != nil {
        return m.LookupFunc(ip)
    }
    return &geoip2.City{}, nil // Default to empty city for tests
}

func (m *MockGeoIPReader) Close() error {
    return nil
}
