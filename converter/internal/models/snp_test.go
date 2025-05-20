package models

import (
	"testing"
)

func TestSubject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		subject Subject
		wantErr bool
		wantMsg string
	}{
		{
			name: "valid genotype",
			subject: Subject{
				Genotype: "AA",
				Match:    "Full",
			},
			wantErr: false,
		},
		{
			name: "invalid genotype (special case)",
			subject: Subject{
				Genotype: "--",
				Match:    "None",
			},
			wantErr: false,
		},
		{
			name: "invalid genotype (empty)",
			subject: Subject{
				Genotype: "",
				Match:    "None",
			},
			wantErr: false,
		},
		{
			name: "invalid genotype (zero)",
			subject: Subject{
				Genotype: "0",
				Match:    "None",
			},
			wantErr: true,
			wantMsg: "invalid genotype format: 0",
		},
		{
			name: "invalid genotype (wrong length)",
			subject: Subject{
				Genotype: "A",
				Match:    "None",
			},
			wantErr: true,
			wantMsg: "invalid genotype format: A",
		},
		{
			name: "invalid nucleotides",
			subject: Subject{
				Genotype: "ZZ",
				Match:    "None",
			},
			wantErr: true,
			wantMsg: "invalid nucleotides in genotype: ZZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.subject.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Subject.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantMsg {
				t.Errorf("Subject.Validate() error message = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestNewSNP(t *testing.T) {
	tests := []struct {
		name     string
		gene     string
		rsid     string
		allele   string
		notes    string
		genotype string
		wantErr  bool
		wantMsg  string
	}{
		{
			name:     "valid SNP",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "AA",
			wantErr:  false,
		},
		{
			name:     "valid SNP with partial match",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "AG",
			wantErr:  false,
		},
		{
			name:     "valid SNP with no match",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "GG",
			wantErr:  false,
		},
		{
			name:     "special case genotype (--)",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "--",
			wantErr:  false,
		},
		{
			name:     "special case genotype (empty)",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "",
			wantErr:  false,
		},
		{
			name:     "invalid genotype (0)",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "0",
			wantErr:  true,
			wantMsg:  "invalid subject: invalid genotype format: 0",
		},
		{
			name:     "invalid genotype format (too short)",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "A",
			wantErr:  true,
			wantMsg:  "invalid subject: invalid genotype format: A",
		},
		{
			name:     "invalid genotype format (invalid nucleotides)",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "ZZ",
			wantErr:  true,
			wantMsg:  "invalid subject: invalid nucleotides in genotype: ZZ",
		},
		{
			name:     "empty gene",
			gene:     "",
			rsid:     "rs123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "AA",
			wantErr:  true,
			wantMsg:  "gene cannot be empty",
		},
		{
			name:     "empty RSID",
			gene:     "TestGene",
			rsid:     "",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "AA",
			wantErr:  true,
			wantMsg:  "RSID cannot be empty",
		},
		{
			name:     "invalid RSID format (no rs prefix)",
			gene:     "TestGene",
			rsid:     "123",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "AA",
			wantErr:  true,
			wantMsg:  "invalid RSID format: 123",
		},
		{
			name:     "invalid RSID format (non-numeric after rs)",
			gene:     "TestGene",
			rsid:     "rsabc",
			allele:   "A",
			notes:    "Test SNP",
			genotype: "AA",
			wantErr:  true,
			wantMsg:  "invalid RSID format: rsabc",
		},
		{
			name:     "empty allele",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "",
			notes:    "Test SNP",
			genotype: "AA",
			wantErr:  true,
			wantMsg:  "allele cannot be empty",
		},
		{
			name:     "invalid allele (not A/T/C/G)",
			gene:     "TestGene",
			rsid:     "rs123",
			allele:   "X",
			notes:    "Test SNP",
			genotype: "AA",
			wantErr:  true,
			wantMsg:  "invalid allele: X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snp, err := NewSNP(tt.gene, tt.rsid, tt.allele, tt.notes, tt.genotype)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSNP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantMsg {
				t.Errorf("NewSNP() error message = %q, want %q", err.Error(), tt.wantMsg)
			}
			if !tt.wantErr && snp != nil {
				// Verify the SNP was created correctly
				if snp.Gene != tt.gene {
					t.Errorf("SNP.Gene = %v, want %v", snp.Gene, tt.gene)
				}
				if snp.RSID != tt.rsid {
					t.Errorf("SNP.RSID = %v, want %v", snp.RSID, tt.rsid)
				}
				if snp.Allele != tt.allele {
					t.Errorf("SNP.Allele = %v, want %v", snp.Allele, tt.allele)
				}
				if snp.Notes != tt.notes {
					t.Errorf("SNP.Notes = %v, want %v", snp.Notes, tt.notes)
				}
				// Verify Subject was created correctly
				if snp.Subject.Genotype != tt.genotype {
					t.Errorf("SNP.Subject.Genotype = %v, want %v", snp.Subject.Genotype, tt.genotype)
				}
				if tt.genotype == "--" || tt.genotype == "" || tt.genotype == "0" {
					if snp.Subject.Match != "None" {
						t.Errorf("SNP.Subject.Match = %v, want %v", snp.Subject.Match, "None")
					}
				} else {
					match := DetermineMatch(tt.genotype, tt.allele)
					if snp.Subject.Match != match {
						t.Errorf("SNP.Subject.Match = %v, want %v", snp.Subject.Match, match)
					}
				}
			}
		})
	}
}

func TestSNP_Validate(t *testing.T) {
	tests := []struct {
		name    string
		snp     SNP
		wantErr bool
		wantMsg string
	}{
		{
			name: "valid SNP",
			snp: SNP{
				Gene:    "TestGene",
				RSID:    "rs123",
				Allele:  "A",
				Notes:   "Test SNP",
				Subject: Subject{Genotype: "AA", Match: "Full"},
			},
			wantErr: false,
		},
		{
			name: "invalid SNP (invalid genotype)",
			snp: SNP{
				Gene:    "TestGene",
				RSID:    "rs123",
				Allele:  "A",
				Notes:   "Test SNP",
				Subject: Subject{Genotype: "ZZ", Match: "None"},
			},
			wantErr: true,
			wantMsg: "invalid subject: invalid nucleotides in genotype: ZZ",
		},
		{
			name: "invalid SNP (special case --)",
			snp: SNP{
				Gene:    "TestGene",
				RSID:    "rs123",
				Allele:  "A",
				Notes:   "Test SNP",
				Subject: Subject{Genotype: "--", Match: "None"},
			},
			wantErr: false,
		},
		{
			name: "invalid SNP (empty genotype)",
			snp: SNP{
				Gene:    "TestGene",
				RSID:    "rs123",
				Allele:  "A",
				Notes:   "Test SNP",
				Subject: Subject{Genotype: "", Match: "None"},
			},
			wantErr: false,
		},
		{
			name: "invalid SNP (invalid nucleotides)",
			snp: SNP{
				Gene:    "TestGene",
				RSID:    "rs123",
				Allele:  "A",
				Notes:   "Test SNP",
				Subject: Subject{Genotype: "ZZ", Match: "None"},
			},
			wantErr: true,
			wantMsg: "invalid subject: invalid nucleotides in genotype: ZZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.snp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SNP.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantMsg {
				t.Errorf("SNP.Validate() error message = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestDetermineMatch(t *testing.T) {
	tests := []struct {
		name     string
		genotype string
		allele   string
		want     string
	}{
		{
			name:     "full match",
			genotype: "AA",
			allele:   "A",
			want:     "Full",
		},
		{
			name:     "partial match",
			genotype: "AG",
			allele:   "A",
			want:     "Partial",
		},
		{
			name:     "no match",
			genotype: "GG",
			allele:   "A",
			want:     "None",
		},
		{
			name:     "invalid genotype",
			genotype: "--",
			allele:   "A",
			want:     "None",
		},
		{
			name:     "empty genotype",
			genotype: "",
			allele:   "A",
			want:     "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetermineMatch(tt.genotype, tt.allele); got != tt.want {
				t.Errorf("DetermineMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
