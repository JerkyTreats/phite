package pipeline

import (
	"context"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/db"
)

func setupTestRepositories(t *testing.T) (db.DBRepository, db.DBRepository) {
	gwasRepo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
		"path": "testdata/gwas.duckdb",
	})
	if err != nil {
		t.Fatalf("failed to create GWAS repository: %v", err)
	}

	refRepo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
		"path": "testdata/reference.duckdb",
	})
	if err != nil {
		t.Fatalf("failed to create reference repository: %v", err)
	}

	return gwasRepo, refRepo
}

func TestRun_SingleTrait_Success(t *testing.T) {
	gwasRepo, refRepo := setupTestRepositories(t)

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		GWASRepository: gwasRepo,
		GWASTable:      "gwas_table",
		RefRepository:  refRepo,
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
		Ancestry:       "EUR",
		Model:          "v1",
	}

	out, err := Run(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(out.TraitSummaries) != 1 {
		t.Errorf("expected 1 trait summary, got %d", len(out.TraitSummaries))
	}

	// Verify the trait summary content
	if len(out.TraitSummaries) > 0 {
		summary := out.TraitSummaries[0]
		if summary.Trait == "" {
			t.Error("expected non-empty trait name in summary")
		}
		if summary.EffectWeightedContribution == 0 {
			t.Error("expected non-zero effect weighted contribution in summary")
		}
	}
}

func TestRun_MultiTrait_Success(t *testing.T) {
	gwasRepo, refRepo := setupTestRepositories(t)

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		GWASRepository: gwasRepo,
		GWASTable:      "gwas_table",
		RefRepository:  refRepo,
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
		Ancestry:       "EUR",
		Model:          "v1",
	}

	out, err := Run(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(out.TraitSummaries) < 2 {
		t.Errorf("expected >=2 trait summaries for multi-trait input, got %d", len(out.TraitSummaries))
	}

	// Verify trait summaries
	traits := make(map[string]bool)
	for _, summary := range out.TraitSummaries {
		traits[summary.Trait] = true
		if summary.EffectWeightedContribution == 0 {
			t.Errorf("expected non-zero effect weighted contribution for trait %s", summary.Trait)
		}
	}
	if len(traits) < 2 {
		t.Error("expected at least 2 different traits in summaries")
	}
}

func TestRun_ErrorOnMissingInput(t *testing.T) {
	input := PipelineInput{}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on missing input, got nil")
	}
}

func TestRun_ErrorOnMissingRepository(t *testing.T) {
	input := PipelineInput{
		GenotypeFile: "testdata/genotype_single_trait.txt",
		SNPs:         []string{"rs1", "rs2"},
		GWASTable:    "gwas_table",
		OutputFormat: "json",
		OutputPath:   "",
		Ancestry:     "EUR",
		Model:        "v1",
	}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on missing repository, got nil")
	}
}

func TestRun_ErrorOnInvalidGenotypeFile(t *testing.T) {
	gwasRepo, refRepo := setupTestRepositories(t)

	input := PipelineInput{
		GenotypeFile:   "testdata/nonexistent.txt",
		SNPs:           []string{"rs1", "rs2"},
		GWASRepository: gwasRepo,
		GWASTable:      "gwas_table",
		RefRepository:  refRepo,
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
		Ancestry:       "EUR",
		Model:          "v1",
	}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on invalid genotype file, got nil")
	}
}
