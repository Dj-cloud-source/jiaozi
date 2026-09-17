package model

import "time"

type AdminAuditLog struct {
	ID           uint64    `db:"id"`
	AdminID      uint64    `db:"admin_id"`
	Action       string    `db:"action"`
	ResourceType string    `db:"resource_type"`
	ResourceID   *string   `db:"resource_id"`
	BeforeData   *string   `db:"before_data"`
	AfterData    *string   `db:"after_data"`
	CreatedAt    time.Time `db:"created_at"`
}
