package infrastructure

import (
	"time"

	"github.com/golang-todo/modular-monolith/internal/core/errors"
	"github.com/golang-todo/modular-monolith/internal/modules/identity/domain"
	"gorm.io/gorm"
)

type UserModel struct {
	ID           string `gorm:"primaryKey;type:varchar(36)"`
	TenantID     string `gorm:"index;type:varchar(64)"`
	Email        string `gorm:"uniqueIndex;type:varchar(191)"`
	PasswordHash string `gorm:"type:text"`
	FullName     string `gorm:"type:varchar(191)"`
	Role         string `gorm:"type:varchar(32)"`
	IsActive     bool   `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserModel) TableName() string {
	return "users"
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) domain.UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Save(user *domain.User) error {
	model := UserModel{
		ID:           user.ID,
		TenantID:     user.TenantID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		Role:         string(user.Role),
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	return r.db.Save(&model).Error
}

func (r *GormUserRepository) FindByID(id string) (*domain.User, error) {
	var m UserModel
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("user not found")
		}
		return nil, err
	}
	return toDomainUser(&m), nil
}

func (r *GormUserRepository) FindByEmail(email string) (*domain.User, error) {
	var m UserModel
	if err := r.db.First(&m, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("user not found")
		}
		return nil, err
	}
	return toDomainUser(&m), nil
}

func (r *GormUserRepository) List(tenantID string) ([]*domain.User, error) {
	var models []UserModel
	query := r.db
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(models))
	for i, m := range models {
		users[i] = toDomainUser(&m)
	}
	return users, nil
}

func toDomainUser(m *UserModel) *domain.User {
	return &domain.User{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		FullName:     m.FullName,
		Role:         domain.Role(m.Role),
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
