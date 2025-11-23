package adaptive

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

// GeoIPReaderImpl implements the GeoIPReader interface using geoip2-golang.
type GeoIPReaderImpl struct {
	db *geoip2.Reader
}

// NewGeoIPReader creates a new GeoIPReaderImpl from a database file.
func NewGeoIPReader(dbPath string) (*GeoIPReaderImpl, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open geoip database: %w", err)
	}
	return &GeoIPReaderImpl{db: db}, nil
}

// City looks up the city for the given IP address.
func (r *GeoIPReaderImpl) City(ipAddress net.IP) (*geoip2.City, error) {
	return r.db.City(ipAddress)
}

// Close closes the database reader.
func (r *GeoIPReaderImpl) Close() error {
	return r.db.Close()
}
