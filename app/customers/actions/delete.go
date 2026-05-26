package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Delete remove permanentemente (hard delete) o customer identificado por customerID.
// Retorna ErrForbidden se requestingRole não for "owner" nem "admin".
// Retorna ErrCustomerNotFound se o ID não existir ou pertencer a outra org.
// Retorna ErrCustomerHasActiveSchedules se houver agendamentos com status "pending" ou "confirmed".
// Agendamentos com status "completed" ou "cancelled" têm customer_id zerado via FK set null.
func Delete(db *gorm.DB, orgSlug string, customerID uint, requestingUserID uint, requestingRole string) error {
	if requestingRole != "owner" && requestingRole != "admin" {
		return ErrForbidden
	}

	var org models.Organization
	if result := db.Where("slug = ?", orgSlug).Find(&org); result.Error != nil || org.ID == 0 {
		return ErrOrgNotFound
	}

	var customer models.Customer
	if result := db.Where("id = ? AND organization_id = ?", customerID, org.ID).Find(&customer); result.Error != nil || customer.ID == 0 {
		return ErrCustomerNotFound
	}

	var activeSchedule models.Schedule
	if result := db.Where("customer_id = ? AND status IN (?)", customerID, []string{"pending", "confirmed"}).Find(&activeSchedule); result.Error == nil && activeSchedule.ID != 0 {
		return ErrCustomerHasActiveSchedules
	}

	return db.Delete(&customer).Error
}
