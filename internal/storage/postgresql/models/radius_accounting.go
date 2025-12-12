package models

import (
	"time"
)

// RadiusAccounting represents a RADIUS accounting record.
type RadiusAccounting struct {
	ID             string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	SessionID      string    `gorm:"type:varchar(128);not null;index" json:"session_id"`
	UserID         string    `gorm:"type:varchar(64);index" json:"user_id"`
	Username       string    `gorm:"type:varchar(256)" json:"username"`
	NASIdentifier  string    `gorm:"type:varchar(128)" json:"nas_identifier"`
	NASIPAddress   string    `gorm:"type:varchar(45)" json:"nas_ip_address"`
	NASPort        int       `gorm:"type:int" json:"nas_port"`
	StatusType     int       `gorm:"type:int;not null" json:"status_type"`
	SessionTime    int       `gorm:"type:int;default:0" json:"session_time"`
	InputOctets    uint64    `gorm:"type:bigint;default:0" json:"input_octets"`
	OutputOctets   uint64    `gorm:"type:bigint;default:0" json:"output_octets"`
	InputPackets   uint64    `gorm:"type:bigint;default:0" json:"input_packets"`
	OutputPackets  uint64    `gorm:"type:bigint;default:0" json:"output_packets"`
	TerminateCause int       `gorm:"type:int" json:"terminate_cause"`
	FramedIP       string    `gorm:"type:varchar(45)" json:"framed_ip"`
	CalledStation  string    `gorm:"type:varchar(128)" json:"called_station"`
	CallingStation string    `gorm:"type:varchar(128)" json:"calling_station"`
	CreatedAt      time.Time `gorm:"default:now();index" json:"created_at"`
}
