package snps

import (
	"testing"
)

func TestCleanAndValidateSNPs(t *testing.T) {
	cases := []struct {
		name    string
		input   []string
		expect  []string
		hasErr  bool
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
