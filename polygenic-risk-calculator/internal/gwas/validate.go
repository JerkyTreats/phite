package gwas

import (
	"context"

	"phite.io/polygenic-risk-calculator/internal/db"
)

// ValidateGWASDBAndTable checks that the database is reachable and the table exists using the repository abstraction.
func ValidateGWASDBAndTable(repo db.DBRepository, table string) error {
	return repo.TestConnection(context.Background(), table)
}
