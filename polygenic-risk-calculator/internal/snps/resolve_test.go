package snps

import (
	"os"
	"testing"
)

func TestResolveSNPs(t *testing.T) {
	// direct input
	out, err := ResolveSNPs([]string{"rs1", "rs2", "rs1"}, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(out) != 2 || out[0] != "rs1" || out[1] != "rs2" {
		t.Errorf("deduplication failed: %v", out)
	}

	// file input (create temp file with supported .csv extension)
	f, err := os.CreateTemp("", "snps_test_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("rs3\nrs4\n")
	f.Close()

	out, err = ResolveSNPs(nil, f.Name())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(out) != 2 || out[0] != "rs3" || out[1] != "rs4" {
		t.Errorf("file parse failed: %v", out)
	}

	// neither provided
	_, err = ResolveSNPs(nil, "")
	if err == nil {
		t.Errorf("expected error for no input")
	}
}
