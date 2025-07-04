package snps

import (
	"os"
	"strings"
	"testing"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

func TestParseSNPsFromFile_JSONValid(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dir := t.TempDir()
	jsonContent := `[
  "rs123",
  "rs456",
  "rs789"
]`
	path := writeTempFile(t, dir, "*.json", jsonContent)

	rsids, err := ParseSNPsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"rs123", "rs456", "rs789"}
	if len(rsids) != len(want) {
		t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
	}
	for i := range want {
		if rsids[i] != want[i] {
			t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
		}
	}
}

func TestParseSNPsFromFile_CSVOnePerLine(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dir := t.TempDir()
	csvContent := "rs123\nrs456\nrs789\n"
	path := writeTempFile(t, dir, "*.csv", csvContent)

	rsids, err := ParseSNPsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"rs123", "rs456", "rs789"}
	if len(rsids) != len(want) {
		t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
	}
	for i := range want {
		if rsids[i] != want[i] {
			t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
		}
	}
}

func TestParseSNPsFromFile_CSVWithHeader(t *testing.T) {
	logging.SetSilentLoggingForTest()
	t.Run("tsv single-column header", func(t *testing.T) {
		dir := t.TempDir()
		tsvContent := "rsid\nrs123\nrs456\nrs789\n"
		path := writeTempFile(t, dir, "*.tsv", tsvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456", "rs789"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("tsv multi-column header, rsid middle", func(t *testing.T) {
		dir := t.TempDir()
		tsvContent := "effect\trsid\tpval\n0.12\trs123\t0.01\n0.08\trs456\t0.02\n"
		path := writeTempFile(t, dir, "*.tsv", tsvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("tsv multi-column header, rsid first", func(t *testing.T) {
		dir := t.TempDir()
		tsvContent := "rsid\teffect\tpval\nrs123\t0.12\t0.01\nrs456\t0.08\t0.02\n"
		path := writeTempFile(t, dir, "*.tsv", tsvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("tsv multi-column header, rsid last", func(t *testing.T) {
		dir := t.TempDir()
		tsvContent := "effect\tpval\trsid\n0.12\t0.01\trs123\n0.08\t0.02\trs456\n"
		path := writeTempFile(t, dir, "*.tsv", tsvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("tsv multi-column header, RSID uppercase", func(t *testing.T) {
		dir := t.TempDir()
		tsvContent := "EFFECT\tRSID\tPVAL\n0.12\trs123\t0.01\n0.08\trs456\t0.02\n"
		path := writeTempFile(t, dir, "*.tsv", tsvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("single-column header", func(t *testing.T) {
		dir := t.TempDir()
		csvContent := "rsid\nrs123\nrs456\nrs789\n"
		path := writeTempFile(t, dir, "*.csv", csvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456", "rs789"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("multi-column header, rsid middle", func(t *testing.T) {
		dir := t.TempDir()
		csvContent := "effect,rsid,pval\n0.12,rs123,0.01\n0.08,rs456,0.02\n"
		path := writeTempFile(t, dir, "*.csv", csvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("multi-column header, rsid first", func(t *testing.T) {
		dir := t.TempDir()
		csvContent := "rsid,effect,pval\nrs123,0.12,0.01\nrs456,0.08,0.02\n"
		path := writeTempFile(t, dir, "*.csv", csvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("multi-column header, rsid last", func(t *testing.T) {
		dir := t.TempDir()
		csvContent := "effect,pval,rsid\n0.12,0.01,rs123\n0.08,0.02,rs456\n"
		path := writeTempFile(t, dir, "*.csv", csvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("multi-column header, RSID uppercase", func(t *testing.T) {
		dir := t.TempDir()
		csvContent := "EFFECT,RSID,PVAL\n0.12,rs123,0.01\n0.08,rs456,0.02\n"
		path := writeTempFile(t, dir, "*.csv", csvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("multi-column header, RS ID with space and mixed case (csv)", func(t *testing.T) {
		dir := t.TempDir()
		csvContent := "EFFECT, RS ID ,PVAL\n0.12,rs123,0.01\n0.08,rs456,0.02\n"
		path := writeTempFile(t, dir, "*.csv", csvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("multi-column header, RS ID with space and mixed case (tsv)", func(t *testing.T) {
		dir := t.TempDir()
		tsvContent := "Group\tGene\tRS ID\tEffect Allele\tYour Genotype\tNotes About Effect Allele\t0\nDiabetes\tMTNR1B\trs10830963\tG\tCC\tIncreased fasting glucose levels, increased risk of type 2 diabetes (2-fold) when eating late at night\n"
		path := writeTempFile(t, dir, "*.tsv", tsvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs10830963"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

	t.Run("multi-column header, RS ID with space and mixed case (tsv)", func(t *testing.T) {
		tsvContent := "EFFECT\t RS ID \tPVAL\n0.12\trs123\t0.01\n0.08\trs456\t0.02\n"
		path := writeTempFile(t, t.TempDir(), "*.tsv", tsvContent)

		rsids, err := ParseSNPsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"rs123", "rs456"}
		if len(rsids) != len(want) {
			t.Fatalf("expected %d rsids, got %d", len(want), len(rsids))
		}
		for i := range want {
			if rsids[i] != want[i] {
				t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
			}
		}
	})

}
func TestParseSNPsFromFile_IgnoreBlankLines(t *testing.T) {
	logging.SetSilentLoggingForTest()
	tests := []struct {
		name    string
		ext     string
		content string
		want    []string
	}{
		{
			name:    "csv blank lines and whitespace",
			ext:     ".csv",
			content: "\n  rs123 \n\nrs456\n \nrs789 \n\n",
			want:    []string{"rs123", "rs456", "rs789"},
		},
		{
			name:    "json blank lines and whitespace",
			ext:     ".json",
			content: "[\n  \"rs123\",\n  \"rs456\",\n  \n  \"rs789\"\n]\n",
			want:    []string{"rs123", "rs456", "rs789"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeTempFile(t, dir, "*"+tc.ext, tc.content)
			rsids, err := ParseSNPsFromFile(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(rsids) != len(tc.want) {
				t.Fatalf("expected %d rsids, got %d", len(tc.want), len(rsids))
			}
			for i := range tc.want {
				if strings.TrimSpace(rsids[i]) != tc.want[i] {
					t.Errorf("expected rsid[%d]=%q, got %q", i, tc.want[i], rsids[i])
				}
			}
		})
	}
}

func TestParseSNPsFromFile_MalformedJSON(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dir := t.TempDir()
	badJSON := "[\"rs123\", \"rs456\"" // missing closing bracket
	path := writeTempFile(t, dir, "*.json", badJSON)
	_, err := ParseSNPsFromFile(path)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestParseSNPsFromFile_MalformedCSV(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dir := t.TempDir()
	badCSV := "rs123\n\x00\nrs456\n" // contains a null byte
	path := writeTempFile(t, dir, "*.csv", badCSV)
	_, err := ParseSNPsFromFile(path)
	if err == nil {
		t.Error("expected error for malformed CSV, got nil")
	}
}

func TestParseSNPsFromFile_UnsupportedExtension(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dir := t.TempDir()
	content := "rs123\nrs456\n"
	path := writeTempFile(t, dir, "*.xls", content)
	_, err := ParseSNPsFromFile(path)
	if err == nil {
		t.Error("expected error for unsupported file extension, got nil")
	}
}

func TestParseSNPsFromFile_DuplicateRSIDs(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dir := t.TempDir()
	csvContent := "rs123\nrs456\nrs123\n"
	path := writeTempFile(t, dir, "*.csv", csvContent)
	rsids, err := ParseSNPsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"rs123", "rs456"}
	if len(rsids) != len(want) {
		t.Fatalf("expected %d unique rsids, got %d", len(want), len(rsids))
	}
	for i := range want {
		if rsids[i] != want[i] {
			t.Errorf("expected rsid[%d]=%q, got %q", i, want[i], rsids[i])
		}
	}
}

func TestParseSNPsFromFile_EmptyOrMissingRSID(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dir := t.TempDir()
	csvContent := "rs123\n\n   \nrs456\n"
	path := writeTempFile(t, dir, "*.csv", csvContent)
	rsids, _ := ParseSNPsFromFile(path)
	for _, r := range rsids {
		if strings.TrimSpace(r) == "" {
			t.Errorf("found empty rsid in output: %q", r)
		}
	}
}

// Helper to create temp files for tests
func writeTempFile(t *testing.T, dir, pattern, content string) string {
	t.Helper()
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}
