package infrastructure

import (
	"time"

	"github.com/golang-todo/modular-monolith/internal/core/errors"
	"github.com/golang-todo/modular-monolith/internal/modules/todo/domain"
	"gorm.io/gorm"
)

type TodoModel struct {
	ID          string     `gorm:"primaryKey;type:varchar(36)"`
	TenantID    string     `gorm:"index;type:varchar(64)"`
	UserID      string     `gorm:"index;type:varchar(64)"`
	Title       string     `gorm:"type:varchar(255)"`
	Description string     `gorm:"type:text"`
	Status      string     `gorm:"index;type:varchar(32)"`
	Priority    string     `gorm:"index;type:varchar(32)"`
	DueDate     *time.Time `gorm:"index"`
	Tags        string     `gorm:"type:varchar(255)"`
	Version     int64      `gorm:"default:1"` // Optimistic Concurrency Control
	CreatedAt   time.Time  `gorm:"index"`
	UpdatedAt   time.Time
}

func (TodoModel) TableName() string {
	return "todos"
}

type GormTodoRepository struct {
	db *gorm.DB
}

func NewGormTodoRepository(db *gorm.DB) domain.TodoRepository {
	return &GormTodoRepository{db: db}
}

func (r *GormTodoRepository) Save(todo *domain.Todo) error {
	m := TodoModel{
		ID:          todo.ID,
		TenantID:    todo.TenantID,
		UserID:      todo.UserID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      string(todo.Status),
		Priority:    string(todo.Priority),
		DueDate:     todo.DueDate,
		Tags:        todo.Tags,
		Version:     todo.Version,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
	return r.db.Create(&m).Error
}

func (r *GormTodoRepository) UpdateWithVersion(todo *domain.Todo, expectedVersion int64) error {
	// Compare-and-Swap (CAS) optimistic concurrency update
	res := r.db.Model(&TodoModel{}).
		Where("id = ? AND tenant_id = ? AND version = ?", todo.ID, todo.TenantID, expectedVersion).
		Updates(map[string]interface{}{
			"title":       todo.Title,
			"description": todo.Description,
			"status":      string(todo.Status),
			"priority":    string(todo.Priority),
			"due_date":    todo.DueDate,
			"tags":        todo.Tags,
			"version":     todo.Version,
			"updated_at":  todo.UpdatedAt,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.NewConflictError("concurrent update conflict: todo was modified by another operation")
	}
	return nil
}

func (r *GormTodoRepository) FindByID(id, tenantID string) (*domain.Todo, error) {
	var m TodoModel
	query := r.db.Where("id = ?", id)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("todo item not found")
		}
		return nil, err
	}
	return toDomainTodo(&m), nil
}

func (r *GormTodoRepository) Delete(id, tenantID string) error {
	res := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&TodoModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.NewNotFoundError("todo item not found")
	}
	return nil
}

func (r *GormTodoRepository) List(
	tenantID, userID string,
	status *domain.TodoStatus,
	priority *domain.TodoPriority,
	search string,
	limit, offset int,
) ([]*domain.Todo, int64, error) {
	var count int64
	var models []TodoModel

	query := r.db.Model(&TodoModel{})
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", string(*status))
	}
	if priority != nil && *priority != "" {
		query = query.Where("priority = ?", string(*priority))
	}
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR tags LIKE ?", s, s, s)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	todos := make([]*domain.Todo, len(models))
	for i, m := range models {
		todos[i] = toDomainTodo(&m)
	}
	return todos, count, nil
}

func (r *GormTodoRepository) GetStats(tenantID, userID string) (map[string]int64, error) {
	type Result struct {
		Status string
		Count  int64
	}
	var results []Result

	query := r.db.Model(&TodoModel{}).Select("status, count(*) as count")
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Group("status").Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := map[string]int64{
		"TOTAL":       0,
		"PENDING":     0,
		"IN_PROGRESS": 0,
		"COMPLETED":   0,
		"CANCELLED":   0,
	}

	for _, res := range results {
		stats[res.Status] = res.Count
		stats["TOTAL"] += res.Count
	}

	return stats, nil
}

func toDomainTodo(m *TodoModel) *domain.Todo {
	return &domain.Todo{
		ID:          m.ID,
		TenantID:    m.TenantID,
		UserID:      m.UserID,
		Title:       m.Title,
		Description: m.Description,
		Status:      domain.TodoStatus(m.Status),
		Priority:    domain.TodoPriority(m.Priority),
		DueDate:     m.DueDate,
		Tags:        m.Tags,
		Version:     m.Version,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
