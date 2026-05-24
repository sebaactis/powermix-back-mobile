package prode

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Constantes de estado de premio
const (
	RewardStatusPending          = "PENDING"
	RewardStatusFulfilled        = "FULFILLED"
	RewardStatusPendingInventory = "PENDING_INVENTORY"
	RewardStatusFailed           = "FAILED"
	RewardStatusSkipped          = "SKIPPED"
)

type ProdeReward struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PredictionID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_prode_reward_pred" json:"prediction_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	VoucherID     *uuid.UUID `gorm:"type:uuid" json:"voucher_id,omitempty"`
	Status        string     `gorm:"type:varchar(30);not null;default:PENDING" json:"status"`
	AdminNotified bool       `gorm:"not null;default:false" json:"admin_notified"`
	FailureReason string     `gorm:"type:varchar(255)" json:"failure_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (ProdeReward) TableName() string { return "prode_rewards" }

func (r *ProdeReward) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
