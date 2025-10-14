package dbconfig

type RepositoryConfig struct {
	Type   string            // "bq", "duckdb"
	Params map[string]string // Dynamic parameters
}

type BigQueryConfig struct {
	ProjectID       string
	DatasetID       string
	BillingProject  string
	CredentialsPath string
	Role            string // "reader", "writer"
}

type DuckDBConfig struct {
	Path string
}
