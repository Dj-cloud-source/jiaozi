package adminaudit

import (
	"context"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

type CreateParams struct {
	AdminID      uint64
	Action       string
	ResourceType string
	ResourceID   *string
	BeforeData   *string
	AfterData    *string
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, params CreateParams) (model.AdminAuditLog, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO admin_audit_logs (
			admin_id,
			action,
			resource_type,
			resource_id,
			before_data,
			after_data
		) VALUES (?, ?, ?, ?, ?, ?)`,
		params.AdminID,
		params.Action,
		params.ResourceType,
		params.ResourceID,
		params.BeforeData,
		params.AfterData,
	)
	if err != nil {
		return model.AdminAuditLog{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.AdminAuditLog{}, err
	}

	return model.AdminAuditLog{
		ID:           uint64(id),
		AdminID:      params.AdminID,
		Action:       params.Action,
		ResourceType: params.ResourceType,
		ResourceID:   params.ResourceID,
		BeforeData:   params.BeforeData,
		AfterData:    params.AfterData,
	}, nil
}

func (r *Repository) List(ctx context.Context, query Query) ([]model.AdminAuditLog, error) {
	var logs []model.AdminAuditLog
	err := r.db.SelectContext(
		ctx,
		&logs,
		`SELECT id,
		        admin_id,
		        action,
		        resource_type,
		        resource_id,
		        CAST(before_data AS CHAR) AS before_data,
		        CAST(after_data AS CHAR) AS after_data,
		        created_at
		 FROM admin_audit_logs
		 ORDER BY created_at DESC, id DESC
		 LIMIT ? OFFSET ?`,
		query.PageSize,
		(query.Page-1)*query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
