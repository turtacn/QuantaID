package radius

import (
	"context"
	"fmt"
	"net"
	"sync"
	"gorm.io/gorm"
	"encoding/json"

	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
)

type RADIUSClient struct {
	ID         string
	Name       string
	IPAddress  string
	Secret     string
	TenantID   string
	Enabled    bool
	VendorType string
	Attributes map[string]interface{}
}

type ClientRepository interface {
	GetByIP(ctx context.Context, ip string) (*RADIUSClient, error)
}

// ClientManager manages NAS clients.
type ClientManager struct {
	db    *gorm.DB
	cache map[string]*RADIUSClient // Simple in-memory cache for now
	mu    sync.RWMutex
}

func NewClientManager(db *gorm.DB) *ClientManager {
	return &ClientManager{
		db:    db,
		cache: make(map[string]*RADIUSClient),
	}
}

// GetByIP retrieves a client by IP address.
func (m *ClientManager) GetByIP(ctx context.Context, ip string) (*RADIUSClient, error) {
	m.mu.RLock()
	if client, ok := m.cache[ip]; ok {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	// Find in DB
	// We need to handle exact match first. CIDR logic is more complex and might need
	// fetching all clients and matching. For this phase, exact match or simple query.

	// Assuming simple exact match for now as per P6-T7 description (can be enhanced).
	// "Supports CIDR" usually implies checking if the request IP falls into the subnet.
	// Implementing full CIDR match properly:
	// We can fetch all enabled clients and check. If the list is large, this is inefficient.
	// Better to use PostgreSQL's inet type and containment operators (<<=), but schema used varchar.
	// We'll fetch all and cache.

	// Optimization: Since we don't have many clients usually, fetching all active clients and caching them is feasible.
	// Or we can query by IP.

	var modelClient models.RadiusClient
	// Try exact match first
	if err := m.db.WithContext(ctx).Where("ip_address = ? AND enabled = ?", ip, true).First(&modelClient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// If not found, check CIDR. This requires iterating.
			// Let's implement a rudimentary CIDR check by loading all clients that have a slash.
			// This might be slow if there are many.
			var clients []models.RadiusClient
			if err := m.db.WithContext(ctx).Where("enabled = ? AND ip_address LIKE ?", true, "%/%").Find(&clients).Error; err != nil {
				return nil, err
			}

			requestIP := net.ParseIP(ip)
			if requestIP == nil {
				return nil, fmt.Errorf("invalid request IP")
			}

			for _, c := range clients {
				_, subnet, err := net.ParseCIDR(c.IPAddress)
				if err == nil && subnet.Contains(requestIP) {
					client := m.toDomain(&c)
					// Cache the result for the specific IP
					m.mu.Lock()
					m.cache[ip] = client
					m.mu.Unlock()
					return client, nil
				}
			}

			return nil, fmt.Errorf("client not found")
		}
		return nil, err
	}

	client := m.toDomain(&modelClient)
	m.mu.Lock()
	m.cache[ip] = client
	m.mu.Unlock()
	return client, nil
}

func (m *ClientManager) toDomain(c *models.RadiusClient) *RADIUSClient {
	var attrs map[string]interface{}
	_ = json.Unmarshal(c.Attributes, &attrs)

	return &RADIUSClient{
		ID:         c.ID,
		Name:       c.Name,
		IPAddress:  c.IPAddress,
		Secret:     c.Secret,
		TenantID:   c.TenantID,
		Enabled:    c.Enabled,
		VendorType: c.VendorType,
		Attributes: attrs,
	}
}

// InvalidateCache clears the cache. Call this on updates.
func (m *ClientManager) InvalidateCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]*RADIUSClient)
}
