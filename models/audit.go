package models

import (
	"time"

	"github.com/go-oauth2/oauth2/utils/uuid"
)

// Audit Log Model
// -----------------------------------------------------------------------------
// Tracks system actions performed by employees/admin.
// Used for activity history and compliance.
type Audit struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ActorID   uuid.UUID `gorm:"type:uuid" json:"actor_id"`
	Action    string    `gorm:"type:varchar(100)" json:"action"`
	Entity    string    `gorm:"type:varchar(100)" json:"entity"`
	EntityID  uint      `json:"entity_id"`
	Metadata  string    `gorm:"type:text" json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}
