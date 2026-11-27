package migrations

import "gorm.io/gorm"

// migration006Up used to insert a sample event ("Distributed Time Allocation"), seven
// sample users and an admin user with no password into every database, production included. It
// is now a no-op: the ID stays so the migration history of existing databases does not change,
// and migration 026 removes the rows it left in databases where it already ran.
func migration006Up(db *gorm.DB) error {
	return nil
}

// migration006Down is a no-op: migration006Up no longer inserts anything.
func migration006Down(db *gorm.DB) error {
	return nil
}
