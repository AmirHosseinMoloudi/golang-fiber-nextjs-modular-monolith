package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	ActionCreated AuditAction = "CREATED"
	ActionUpdated AuditAction = "UPDATED"
	ActionDeleted AuditAction = "DELETED"
	ActionStatusChanged AuditAction = "STATUS_CHANGED"
	ActionLogin   AuditAction = "LOGIN"
)

type AuditRecord struct {
	ID           string      `json:"id"`
	TenantID     string      `json:"tenant_id"`
	ActorID      string      `json:"actor_id"`
	ActorEmail   string      `json:"actor_email"`
	Action       AuditAction `json:"action"`
	EntityType   string      `json:"entity_type"`
	EntityID     string      `json:"entity_id"`
	Payload      string      `json:"payload"`
	PreviousHash string      `json:"previous_hash"`
	RecordHash   string      `json:"record_hash"`
	Timestamp    time.Time   `json:"timestamp"`
}

func CalculateHash(prevHash string, ts time.Time, action AuditAction, entityType, entityID, actorID, payload string) string {
	raw := fmt.Sprintf("%s|%d|%s|%s|%s|%s|%s", prevHash, ts.UnixNano(), action, entityType, entityID, actorID, payload)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func NewAuditRecord(tenantID, actorID, actorEmail string, action AuditAction, entityType, entityID, payload, prevHash string) *AuditRecord {
	ts := time.Now().UTC()
	hash := CalculateHash(prevHash, ts, action, entityType, entityID, actorID, payload)

	return &AuditRecord{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		ActorID:      actorID,
		ActorEmail:   actorEmail,
		Action:       action,
		EntityType:   entityType,
		EntityID:     entityID,
		Payload:      payload,
		PreviousHash: prevHash,
		RecordHash:   hash,
		Timestamp:    ts,
	}
}

type AuditRepository interface {
	Append(record *AuditRecord) error
	GetLatestRecord(tenantID string) (*AuditRecord, error)
	List(tenantID string, limit, offset int) ([]*AuditRecord, int64, error)
	VerifyChainIntegrity(tenantID string) (bool, error)
}
