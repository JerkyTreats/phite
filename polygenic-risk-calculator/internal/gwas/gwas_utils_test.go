package gwas_test

import (
	"reflect"
	"testing"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/gwas"
)

func TestMapToGWASList(t *testing.T) {
	m := map[string]model.GWASSNPRecord{
		"rs1": {RSID: "rs1", RiskAllele: "A", Beta: 1.1, Trait: "Trait1"},
		"rs2": {RSID: "rs2", RiskAllele: "G", Beta: -0.5, Trait: "Trait2"},
	}
	list := gwas.MapToGWASList(m)
	if len(list) != 2 {
		t.Errorf("expected 2 records, got %d", len(list))
	}
	found := map[string]bool{"rs1": false, "rs2": false}
	for _, rec := range list {
		if _, ok := found[rec.RSID]; ok {
			found[rec.RSID] = true
		}
	}
	for k, v := range found {
		if !v {
			t.Errorf("missing record for %s", k)
		}
	}

	// Test nil and empty map
	if out := gwas.MapToGWASList(nil); out != nil {
		t.Errorf("expected nil for nil input, got %v", out)
	}
	if out := gwas.MapToGWASList(map[string]model.GWASSNPRecord{}); !reflect.DeepEqual(out, []model.GWASSNPRecord{}) {
		t.Errorf("expected empty slice for empty map, got %v", out)
	}
}
