package gwas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
)

type GWASService struct {
	repo dbinterface.Repository
}

func NewGWASService() *GWASService {
	repo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{"path": config.GetString("gwas_db_path")})
	if err != nil {
		logging.Error("Failed to create GWASService: %v", err)
		return nil
	}
	return &GWASService{
		repo: repo,
	}
}

// FetchGWASRecordsWithTable loads GWAS SNP records for the given rsids from the specified table using the repository abstraction.
func (s *GWASService) FetchGWASRecordsWithTable(ctx context.Context, table string, rsids []string) (map[string]model.GWASSNPRecord, error) {
	if len(rsids) == 0 {
		return map[string]model.GWASSNPRecord{}, nil
	}
	placeholders := make([]string, len(rsids))
	args := make([]interface{}, len(rsids))
	for i, rsid := range rsids {
		placeholders[i] = "?"
		args[i] = rsid
	}
	query := "SELECT rsid, risk_allele, beta, trait FROM " + table + " WHERE rsid IN (" + strings.Join(placeholders, ",") + ")"
	logging.Info("Executing GWAS query for %d SNPs", len(rsids))

	results, err := s.repo.Query(ctx, query, args...)
	if err != nil {
		logging.Error("GWAS query failed: %v", err)
		return nil, err
	}

	recordMap := make(map[string]model.GWASSNPRecord, len(results))
	for _, row := range results {
		rec := model.GWASSNPRecord{
			RSID:       toString(row["rsid"]),
			RiskAllele: toString(row["risk_allele"]),
			Beta:       toFloat64(row["beta"]),
			Trait:      toString(row["trait"]),
		}
		recordMap[rec.RSID] = rec
	}
	logging.Info("Loaded %d GWAS records from DB", len(recordMap))
	return recordMap, nil
}

// FetchGWASRecords loads GWAS SNP records for the given rsids from the configured table using the repository abstraction.
func (s *GWASService) FetchGWASRecords(ctx context.Context, rsids []string) (map[string]model.GWASSNPRecord, error) {
	table := config.GetString("gwas_table")
	if table == "" {
		return nil, fmt.Errorf("gwas_table is not set")
	}
	return s.FetchGWASRecordsWithTable(ctx, table, rsids)
}

// Helper functions for type conversion
func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func toFloat64(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}
