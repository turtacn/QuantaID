package adaptive

import (
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/mock"
)

type MockGeoIPReader struct {
	mock.Mock
}

func (m *MockGeoIPReader) City(ipAddress net.IP) (*geoip2.City, error) {
	args := m.Called(ipAddress)
	return args.Get(0).(*geoip2.City), args.Error(1)
}

func (m *MockGeoIPReader) Close() error {
	args := m.Called()
	return args.Error(0)
}
