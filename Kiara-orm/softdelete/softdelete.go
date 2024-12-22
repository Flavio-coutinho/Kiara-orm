package softdelete

import (
	"time"
)

// SoftDelete representa os campos necessários para soft delete
type SoftDelete struct {
	DeletedAt *time.Time `db:"deleted_at"`
}

// IsDeleted verifica se o registro foi deletado
func (sd *SoftDelete) IsDeleted() bool {
	return sd.DeletedAt != nil
}

// Delete marca o registro como deletado
func (sd *SoftDelete) Delete() {
	now := time.Now()
	sd.DeletedAt = &now
}

// Restore restaura um registro deletado
func (sd *SoftDelete) Restore() {
	sd.DeletedAt = nil
} 