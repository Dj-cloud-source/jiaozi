package adminaudit

import (
	"context"
	"errors"
	"testing"

	"jiaozi/internal/model"
)

type fakeRepository struct {
	createParams CreateParams
	query        Query
}

func (r *fakeRepository) Create(ctx context.Context, params CreateParams) (model.AdminAuditLog, error) {
	r.createParams = params
	return model.AdminAuditLog{
		ID:           1,
		AdminID:      params.AdminID,
		Action:       params.Action,
		ResourceType: params.ResourceType,
		ResourceID:   params.ResourceID,
	}, nil
}

func (r *fakeRepository) List(ctx context.Context, query Query) ([]model.AdminAuditLog, error) {
	r.query = query
	return []model.AdminAuditLog{{ID: 1, AdminID: 1, Action: "CREATE_STATION", ResourceType: "STATION"}}, nil
}

func TestCreateRejectsInvalidAuditLog(t *testing.T) {
	_, err := NewService(&fakeRepository{}).Create(context.Background(), CreateRequest{
		AdminID:      1,
		Action:       "",
		ResourceType: "STATION",
	})
	if !errors.Is(err, ErrInvalidAuditLog) {
		t.Fatalf("expected invalid audit log error, got %v", err)
	}
}

func TestCreateTrimsActionAndResourceType(t *testing.T) {
	repository := &fakeRepository{}
	_, err := NewService(repository).Create(context.Background(), CreateRequest{
		AdminID:      1,
		Action:       " CREATE_STATION ",
		ResourceType: " STATION ",
	})
	if err != nil {
		t.Fatalf("create audit log failed: %v", err)
	}
	if repository.createParams.Action != "CREATE_STATION" {
		t.Fatalf("unexpected action: %s", repository.createParams.Action)
	}
	if repository.createParams.ResourceType != "STATION" {
		t.Fatalf("unexpected resource type: %s", repository.createParams.ResourceType)
	}
}

func TestListNormalizesPagination(t *testing.T) {
	repository := &fakeRepository{}
	_, err := NewService(repository).List(context.Background(), Query{Page: 0, PageSize: 200})
	if err != nil {
		t.Fatalf("list audit logs failed: %v", err)
	}
	if repository.query.Page != 1 {
		t.Fatalf("unexpected page: %d", repository.query.Page)
	}
	if repository.query.PageSize != 100 {
		t.Fatalf("unexpected page size: %d", repository.query.PageSize)
	}
}
