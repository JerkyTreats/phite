package gwas

import (
	"context"

	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
)

// ValidateGWASDBAndTable checks that the database is reachable and the table exists using the repository abstraction.
func ValidateGWASDBAndTable(repo dbinterface.Repository, table string) error {
	return repo.TestConnection(context.Background(), table)
}
