package adminaudit

import (
	"time"

	"jiaozi/internal/model"
)

type Response struct {
	ID           uint64  `json:"id"`
	AdminID      uint64  `json:"admin_id"`
	Action       string  `json:"action"`
	ResourceType string  `json:"resource_type"`
	ResourceID   *string `json:"resource_id"`
	CreatedAt    string  `json:"created_at"`
}

func NewListResponse(logs []model.AdminAuditLog) []Response {
	response := make([]Response, 0, len(logs))
	for _, log := range logs {
		response = append(response, Response{
			ID:           log.ID,
			AdminID:      log.AdminID,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			CreatedAt:    formatTime(log.CreatedAt),
		})
	}

	return response
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02T15:04:05-07:00")
}
