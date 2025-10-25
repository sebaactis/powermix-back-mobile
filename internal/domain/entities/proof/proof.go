package proof

import (
	"time"

	"github.com/google/uuid"
)

type Proof struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    uuid.UUID
	ProofMPID string
	ProofDate time.Time
}
