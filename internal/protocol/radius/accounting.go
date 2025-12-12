package radius

import (
	"context"
	"encoding/binary"
	"time"
	"fmt"
	"net"

	"gorm.io/gorm"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
)

type AccountingHandler struct {
	db *gorm.DB
}

func NewAccountingHandler(db *gorm.DB) *AccountingHandler {
	return &AccountingHandler{db: db}
}

func (h *AccountingHandler) Handle(ctx context.Context, request *Packet, client *RADIUSClient, remoteAddr *net.UDPAddr) (*Packet, error) {
	statusTypeAttr := request.GetAttribute(AttrAcctStatusType)
	if statusTypeAttr == nil {
		return nil, fmt.Errorf("missing Acct-Status-Type")
	}

	if len(statusTypeAttr.Value) != 4 {
		return nil, fmt.Errorf("invalid Acct-Status-Type length")
	}
	statusType := int(binary.BigEndian.Uint32(statusTypeAttr.Value))

	// Common fields
	sessionID := request.GetString(AttrAcctSessionId)
	if sessionID == "" {
		// Log warning but maybe proceed? RFC says SHOULD contain.
		// We'll require it for correlation.
		return nil, fmt.Errorf("missing Acct-Session-Id")
	}

	username := request.GetString(AttrUserName)

	// Create record object
	record := models.RadiusAccounting{
		ID:            fmt.Sprintf("%s-%d", sessionID, time.Now().UnixNano()), // Simple ID generation
		SessionID:     sessionID,
		Username:      username,
		StatusType:    statusType,
		NASIdentifier: request.GetString(AttrNASIdentifier),
		NASIPAddress:  request.GetString(AttrNASIPAddress), // Or use client IP
		CreatedAt:     time.Now(),
	}

	if record.NASIPAddress == "" && remoteAddr != nil {
		record.NASIPAddress = remoteAddr.IP.String()
	}

	// Parse other accounting metrics
	if val := request.GetAttribute(AttrAcctSessionTime); val != nil {
		record.SessionTime = int(binary.BigEndian.Uint32(val.Value))
	}
	if val := request.GetAttribute(AttrAcctInputOctets); val != nil {
		record.InputOctets = uint64(binary.BigEndian.Uint32(val.Value))
	}
	if val := request.GetAttribute(AttrAcctOutputOctets); val != nil {
		record.OutputOctets = uint64(binary.BigEndian.Uint32(val.Value))
	}
	if val := request.GetAttribute(AttrAcctInputPackets); val != nil {
		record.InputPackets = uint64(binary.BigEndian.Uint32(val.Value))
	}
	if val := request.GetAttribute(AttrAcctOutputPackets); val != nil {
		record.OutputPackets = uint64(binary.BigEndian.Uint32(val.Value))
	}
	if val := request.GetAttribute(AttrAcctTerminateCause); val != nil {
		record.TerminateCause = int(binary.BigEndian.Uint32(val.Value))
	}

	// Handle logic based on status type
	switch statusType {
	case AcctStatusStart:
		// Create new session record
		// We might just log every packet, or update a "current sessions" table.
		// For now, we append to log (models.RadiusAccounting is an append-only log essentially).
	case AcctStatusStop:
		// Log stop
	case AcctStatusInterimUpdate:
		// Log update
	case AcctStatusAccountingOn, AcctStatusAccountingOff:
		// System events
	}

	// Save to DB
	// Note: In high throughput, we might want to buffer this or put in a queue.
	// Direct DB write for now.
	if err := h.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("failed to save accounting record: %w", err)
	}

	// Create Response
	response := request.CreateResponse(CodeAccountingResponse)
	return response, nil
}
