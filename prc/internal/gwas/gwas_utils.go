package gwas

import "phite.io/polygenic-risk-calculator/internal/model"

// MapToGWASList converts a map of GWASSNPRecord to a slice for annotation.
func MapToGWASList(m map[string]model.GWASSNPRecord) []model.GWASSNPRecord {
	if m == nil {
		return nil
	}
	records := make([]model.GWASSNPRecord, 0, len(m))
	for _, rec := range m {
		records = append(records, rec)
	}
	return records
}
