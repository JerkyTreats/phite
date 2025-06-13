package cli

import (
	"context"
	"os"
	"testing"
)

type mockRepo struct{}

func (m *mockRepo) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockRepo) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	return nil
}

func (m *mockRepo) TestConnection(ctx context.Context, table string) error {
	return nil
}

func (m *mockRepo) ValidateTable(ctx context.Context, table string, requiredColumns []string) error {
	return nil
}

func TestParseOptions_CLIOnly(t *testing.T) {
	args := []string{
		"--genotype-file", "geno.txt",
		"--snps", "rs1,rs2,rs3",
		"--gwas-db", "gwas.db",
	}
	opts, err := ParseOptions(args, &mockRepo{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.GenotypeFile != "geno.txt" {
		t.Errorf("expected genotype-file=geno.txt, got %q", opts.GenotypeFile)
	}
	if len(opts.SNPs) != 3 || opts.SNPs[0] != "rs1" {
		t.Errorf("expected SNPs parsed correctly, got %v", opts.SNPs)
	}
	if opts.GWASDB != "gwas.db" {
		t.Errorf("expected gwas-db=gwas.db, got %q", opts.GWASDB)
	}
}

func TestParseOptions_MissingRequired(t *testing.T) {
	args := []string{}
	_, err := ParseOptions(args, &mockRepo{})
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
}

func TestParseOptions_SnpsMutualExclusion(t *testing.T) {
	args := []string{
		"--genotype-file", "geno.txt",
		"--snps", "rs1,rs2",
		"--snps-file", "snps.txt",
		"--gwas-db", "gwas.db",
	}
	_, err := ParseOptions(args, &mockRepo{})
	if err == nil || err.Error() == "" || err.Error() == "<nil>" {
		t.Fatal("expected error for mutually exclusive snps and snps-file")
	}
}

func TestParseOptions_EnvAndConfigFallback(t *testing.T) {
	os.Setenv("GWAS_DUCKDB", "env_gwas.db")
	os.Setenv("GENOTYPE_FILE", "env_geno.txt")
	defer os.Unsetenv("GWAS_DUCKDB")
	defer os.Unsetenv("GENOTYPE_FILE")

	args := []string{
		"--snps", "rs1",
	}
	// This test assumes config package will check env vars properly.
	// If config.GetString("gwas_db_path") checks GWAS_DUCKDB env, this will work.
	opts, err := ParseOptions(args, &mockRepo{})
	if err == nil {
		// Should error because genotype-file is still required
		t.Errorf("expected error for missing genotype-file, got opts: %+v", opts)
	}
}
