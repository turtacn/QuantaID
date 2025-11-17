package adaptive

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoIPReader interface {
	City(ipAddress net.IP) (*geoip2.City, error)
	Close() error
}
