package pipeline

import (
	"testing"
)

func TestRun_SingleTrait_Success(t *testing.T) {
	input := PipelineInput{
		GenotypeFile: "testdata/genotype_single_trait.txt",
		SNPs:         []string{"rs1", "rs2"},
		GWASDB:       "testdata/gwas.duckdb",
		GWASTable:    "gwas_table",
		ReferenceDB:  "testdata/reference.duckdb",
		OutputFormat: "json",
		OutputPath:   "",
		Ancestry:     "EUR",
		Model:        "v1",
	}

	out, err := Run(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(out.TraitSummaries) != 1 {
		t.Errorf("expected 1 trait summary, got %d", len(out.TraitSummaries))
	}
}

func TestRun_MultiTrait_Success(t *testing.T) {
	input := PipelineInput{
		GenotypeFile: "testdata/genotype_multi_trait.txt",
		SNPs:         []string{"rs1", "rs2", "rs3"},
		GWASDB:       "testdata/gwas.duckdb",
		GWASTable:    "gwas_table",
		ReferenceDB:  "testdata/reference.duckdb",
		OutputFormat: "json",
		OutputPath:   "",
		Ancestry:     "EUR",
		Model:        "v1",
	}

	out, err := Run(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(out.TraitSummaries) < 2 {
		t.Errorf("expected >=2 trait summaries for multi-trait input, got %d", len(out.TraitSummaries))
	}
}

func TestRun_ErrorOnMissingInput(t *testing.T) {
	input := PipelineInput{}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on missing input, got nil")
	}
}
