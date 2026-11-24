package infrastructure

import (
	"time"

	"github.com/golang-todo/modular-monolith/internal/modules/audit/domain"
	"gorm.io/gorm"
)

type AuditRecordModel struct {
	ID           string `gorm:"primaryKey;type:varchar(36)"`
	TenantID     string `gorm:"index;type:varchar(64)"`
	ActorID      string `gorm:"index;type:varchar(64)"`
	ActorEmail   string `gorm:"type:varchar(191)"`
	Action       string `gorm:"type:varchar(64)"`
	EntityType   string `gorm:"index;type:varchar(64)"`
	EntityID     string `gorm:"index;type:varchar(64)"`
	Payload      string `gorm:"type:text"`
	PreviousHash string `gorm:"type:varchar(64)"`
	RecordHash   string `gorm:"type:varchar(64)"`
	Timestamp    time.Time `gorm:"index"`
}

func (AuditRecordModel) TableName() string {
	return "audit_records"
}

type GormAuditRepository struct {
	db *gorm.DB
}

func NewGormAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &GormAuditRepository{db: db}
}

func (r *GormAuditRepository) Append(record *domain.AuditRecord) error {
	m := AuditRecordModel{
		ID:           record.ID,
		TenantID:     record.TenantID,
		ActorID:      record.ActorID,
		ActorEmail:   record.ActorEmail,
		Action:       string(record.Action),
		EntityType:   record.EntityType,
		EntityID:     record.EntityID,
		Payload:      record.Payload,
		PreviousHash: record.PreviousHash,
		RecordHash:   record.RecordHash,
		Timestamp:    record.Timestamp,
	}
	return r.db.Create(&m).Error
}

func (r *GormAuditRepository) GetLatestRecord(tenantID string) (*domain.AuditRecord, error) {
	var m AuditRecordModel
	query := r.db.Order("timestamp desc")
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&m).Error; err != nil {
		return nil, err
	}
	return toDomainAudit(&m), nil
}

func (r *GormAuditRepository) List(tenantID string, limit, offset int) ([]*domain.AuditRecord, int64, error) {
	var count int64
	var models []AuditRecordModel

	query := r.db.Model(&AuditRecordModel{})
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("timestamp desc").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	records := make([]*domain.AuditRecord, len(models))
	for i, m := range models {
		records[i] = toDomainAudit(&m)
	}
	return records, count, nil
}

func (r *GormAuditRepository) VerifyChainIntegrity(tenantID string) (bool, error) {
	var models []AuditRecordModel
	query := r.db.Order("timestamp asc")
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Find(&models).Error; err != nil {
		return false, err
	}

	for i := range models {
		expectedPrev := "0000000000000000000000000000000000000000000000000000000000000000"
		if i > 0 {
			expectedPrev = models[i-1].RecordHash
		}
		if models[i].PreviousHash != expectedPrev {
			return false, nil
		}
		calc := domain.CalculateHash(
			models[i].PreviousHash,
			models[i].Timestamp,
			domain.AuditAction(models[i].Action),
			models[i].EntityType,
			models[i].EntityID,
			models[i].ActorID,
			models[i].Payload,
		)
		if calc != models[i].RecordHash {
			return false, nil
		}
	}
	return true, nil
}

func toDomainAudit(m *AuditRecordModel) *domain.AuditRecord {
	return &domain.AuditRecord{
		ID:           m.ID,
		TenantID:     m.TenantID,
		ActorID:      m.ActorID,
		ActorEmail:   m.ActorEmail,
		Action:       domain.AuditAction(m.Action),
		EntityType:   m.EntityType,
		EntityID:     m.EntityID,
		Payload:      m.Payload,
		PreviousHash: m.PreviousHash,
		RecordHash:   m.RecordHash,
		Timestamp:    m.Timestamp,
	}
}
