package device

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

// DeviceFingerprinter handles device fingerprint generation and validation
type DeviceFingerprinter struct{}

// NewDeviceFingerprinter creates a new DeviceFingerprinter
func NewDeviceFingerprinter() *DeviceFingerprinter {
	return &DeviceFingerprinter{}
}

// GenerateHash generates a consistent hash from fingerprint data
// It normalizes the data to ensure the same input always produces the same hash
func (df *DeviceFingerprinter) GenerateHash(data map[string]interface{}) string {
	// 1. Create a copy to avoid modifying original
	normalized := make(map[string]interface{})

	// 2. Select stable fields for hashing
	// We only use specific fields that are less likely to change frequently or are critical for identification
	stableFields := []string{
		"user_agent",
		"screen_resolution",
		"timezone",
		"language",
		"platform",
		"cpu_class",
		"hardware_concurrency",
		"device_memory",
		"canvas_hash",
		"webgl_hash",
	}

	for _, field := range stableFields {
		if val, ok := data[field]; ok {
			normalized[field] = val
		}
	}

	// 3. Serialize to canonical JSON
	// We sort keys to ensure consistency
	keys := make([]string, 0, len(normalized))
	for k := range normalized {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(":")
		// Simple string representation for now, could be more robust
		valBytes, _ := json.Marshal(normalized[k])
		sb.Write(valBytes)
		sb.WriteString(";")
	}

	// 4. Hash
	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// Compare checks similarity between two raw fingerprints
// Returns a score between 0.0 and 1.0
func (df *DeviceFingerprinter) Compare(oldData, newData map[string]interface{}) float64 {
	if len(oldData) == 0 || len(newData) == 0 {
		return 0.0
	}

	matches := 0
	totalChecks := 0

	// List of fields to check
	fields := []string{
		"user_agent",
		"screen_resolution",
		"timezone",
		"language",
		"platform",
		"cpu_class",
		"hardware_concurrency",
		"device_memory",
		"canvas_hash",
		"webgl_hash",
		"plugins",
		"fonts",
	}

	for _, field := range fields {
		oldVal, oldOk := oldData[field]
		newVal, newOk := newData[field]

		if oldOk {
			totalChecks++
			if newOk && isValueEqual(oldVal, newVal) {
				matches++
			}
		}
	}

	if totalChecks == 0 {
		return 0.0
	}

	return float64(matches) / float64(totalChecks)
}

func isValueEqual(v1, v2 interface{}) bool {
	// Simple equality for basic types
	// For complex types like arrays/maps, we might need deep comparison
	// But fingerprints usually contain strings or numbers
	b1, _ := json.Marshal(v1)
	b2, _ := json.Marshal(v2)
	return string(b1) == string(b2)
}
