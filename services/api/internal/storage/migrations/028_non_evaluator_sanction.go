package migrations

import "gorm.io/gorm"

// migration028Up adds the per-event sanction for participants who do not complete their
// ranking: exclusion from the adjusted ranking (Merrifield & Saari 2009, the default) or a
// penalty of non_evaluator_penalty places. Existing configurations take the default.
func migration028Up(db *gorm.DB) error {
	stmts := []string{
		`ALTER TABLE voting_configurations
			ADD COLUMN IF NOT EXISTS non_evaluator_sanction VARCHAR(20) NOT NULL DEFAULT 'exclusion'`,
		`ALTER TABLE voting_configurations
			ADD COLUMN IF NOT EXISTS non_evaluator_penalty INTEGER NOT NULL DEFAULT 10`,
		`ALTER TABLE voting_configurations DROP CONSTRAINT IF EXISTS valid_non_evaluator_sanction`,
		`ALTER TABLE voting_configurations ADD CONSTRAINT valid_non_evaluator_sanction
			CHECK (non_evaluator_sanction IN ('exclusion', 'penalty'))`,
		`ALTER TABLE voting_configurations DROP CONSTRAINT IF EXISTS valid_non_evaluator_penalty`,
		`ALTER TABLE voting_configurations ADD CONSTRAINT valid_non_evaluator_penalty
			CHECK (non_evaluator_penalty BETWEEN 1 AND 100)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}

func migration028Down(db *gorm.DB) error {
	stmts := []string{
		`ALTER TABLE voting_configurations DROP CONSTRAINT IF EXISTS valid_non_evaluator_penalty`,
		`ALTER TABLE voting_configurations DROP CONSTRAINT IF EXISTS valid_non_evaluator_sanction`,
		`ALTER TABLE voting_configurations DROP COLUMN IF EXISTS non_evaluator_penalty`,
		`ALTER TABLE voting_configurations DROP COLUMN IF EXISTS non_evaluator_sanction`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}
