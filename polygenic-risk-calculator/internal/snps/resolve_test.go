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

func TestCleanAndValidateSNPs(t *testing.T) {
	cases := []struct {
		name   string
		input  []string
		expect []string
		hasErr bool
	}{
		{"empty list", []string{}, nil, true},
		{"only empty strings", []string{"", "  "}, nil, true},
		{"dedup and trim", []string{"rs1", " rs2 ", "rs1", "rs3"}, []string{"rs1", "rs2", "rs3"}, false},
		{"contains empty", []string{"rs1", "", "rs2"}, nil, true},
	}
	for _, c := range cases {
		out, err := CleanAndValidateSNPs(c.input)
		if c.hasErr {
			if err == nil {
				t.Errorf("%s: expected error, got nil", c.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", c.name, err)
			continue
		}
		if len(out) != len(c.expect) {
			t.Errorf("%s: expected %d, got %d", c.name, len(c.expect), len(out))
			continue
		}
		for i := range out {
			if out[i] != c.expect[i] {
				t.Errorf("%s: at %d, expected %q, got %q", c.name, i, c.expect[i], out[i])
			}
		}
	}
}
