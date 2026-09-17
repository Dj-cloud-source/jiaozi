package adminaudit

import (
	"context"
	"strings"

	"jiaozi/internal/model"
)

type repository interface {
	Create(ctx context.Context, params CreateParams) (model.AdminAuditLog, error)
	List(ctx context.Context, query Query) ([]model.AdminAuditLog, error)
}

type Service struct {
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

type CreateRequest struct {
	AdminID      uint64
	Action       string
	ResourceType string
	ResourceID   *string
	BeforeData   *string
	AfterData    *string
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (model.AdminAuditLog, error) {
	req.Action = strings.TrimSpace(req.Action)
	req.ResourceType = strings.TrimSpace(req.ResourceType)
	if req.AdminID == 0 || req.Action == "" || req.ResourceType == "" {
		return model.AdminAuditLog{}, ErrInvalidAuditLog
	}

	return s.repository.Create(ctx, CreateParams{
		AdminID:      req.AdminID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		BeforeData:   req.BeforeData,
		AfterData:    req.AfterData,
	})
}

func (s *Service) List(ctx context.Context, query Query) ([]model.AdminAuditLog, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	return s.repository.List(ctx, query)
}
