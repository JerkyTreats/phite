package reference_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPRSModel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		model   PRSModel
		wantErr bool
	}{
		{
			name: "valid model",
			model: PRSModel{
				ID:    "test_model",
				Trait: "test_trait",
				Variants: []Variant{
					{
						ID:           "rs123",
						Chromosome:   "1",
						Position:     1000,
						EffectAllele: "A",
						OtherAllele:  "G",
						EffectWeight: 0.5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty model ID",
			model: PRSModel{
				ID:    "",
				Trait: "test_trait",
				Variants: []Variant{
					{
						ID:           "rs123",
						Chromosome:   "1",
						Position:     1000,
						EffectAllele: "A",
						OtherAllele:  "G",
						EffectWeight: 0.5,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty variants",
			model: PRSModel{
				ID:       "test_model",
				Trait:    "test_trait",
				Variants: []Variant{},
			},
			wantErr: true,
		},
		{
			name: "duplicate variant IDs",
			model: PRSModel{
				ID:    "test_model",
				Trait: "test_trait",
				Variants: []Variant{
					{
						ID:           "rs123",
						Chromosome:   "1",
						Position:     1000,
						EffectAllele: "A",
						OtherAllele:  "G",
						EffectWeight: 0.5,
					},
					{
						ID:           "rs123",
						Chromosome:   "1",
						Position:     1001,
						EffectAllele: "T",
						OtherAllele:  "C",
						EffectWeight: 0.3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty variant ID",
			model: PRSModel{
				ID:    "test_model",
				Trait: "test_trait",
				Variants: []Variant{
					{
						ID:           "",
						Chromosome:   "1",
						Position:     1000,
						EffectAllele: "A",
						OtherAllele:  "G",
						EffectWeight: 0.5,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPRSModel_GetEffectSizes(t *testing.T) {
	model := PRSModel{
		ID:    "test_model",
		Trait: "test_trait",
		Variants: []Variant{
			{
				ID:           "rs123",
				EffectWeight: 0.5,
			},
			{
				ID:           "rs456",
				EffectWeight: -0.3,
			},
			{
				ID:           "rs789",
				EffectWeight: 0.0,
			},
		},
	}

	effects := model.GetEffectSizes()
	assert.Equal(t, 3, len(effects))
	assert.Equal(t, 0.5, effects["rs123"])
	assert.Equal(t, -0.3, effects["rs456"])
	assert.Equal(t, 0.0, effects["rs789"])
}

func TestFormatVariantID(t *testing.T) {
	tests := []struct {
		name     string
		chrom    string
		pos      int64
		ref      string
		alt      string
		expected string
	}{
		{
			name:     "standard variant",
			chrom:    "1",
			pos:      1000,
			ref:      "A",
			alt:      "G",
			expected: "1:1000:A:G",
		},
		{
			name:     "multi-base variant",
			chrom:    "X",
			pos:      12345,
			ref:      "AT",
			alt:      "GC",
			expected: "X:12345:AT:GC",
		},
		{
			name:     "deletion",
			chrom:    "2",
			pos:      5000,
			ref:      "AT",
			alt:      "A",
			expected: "2:5000:AT:A",
		},
		{
			name:     "insertion",
			chrom:    "3",
			pos:      7500,
			ref:      "A",
			alt:      "AT",
			expected: "3:7500:A:AT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatVariantID(tt.chrom, tt.pos, tt.ref, tt.alt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseVariantID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantChrom string
		wantPos   int64
		wantRef   string
		wantAlt   string
		wantErr   bool
	}{
		{
			name:      "valid variant ID",
			id:        "1:1000:A:G",
			wantChrom: "1",
			wantPos:   1000,
			wantRef:   "A",
			wantAlt:   "G",
			wantErr:   false,
		},
		{
			name:    "invalid format - missing parts",
			id:      "1:1000:A",
			wantErr: true,
		},
		{
			name:    "invalid format - extra parts",
			id:      "1:1000:A:G:extra",
			wantErr: true,
		},
		{
			name:    "invalid position",
			id:      "1:abc:A:G",
			wantErr: true,
		},
		{
			name:      "multi-base variant",
			id:        "X:12345:AT:GC",
			wantChrom: "X",
			wantPos:   12345,
			wantRef:   "AT",
			wantAlt:   "GC",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chrom, pos, ref, alt, err := ParseVariantID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantChrom, chrom)
				assert.Equal(t, tt.wantPos, pos)
				assert.Equal(t, tt.wantRef, ref)
				assert.Equal(t, tt.wantAlt, alt)
			}
		})
	}
}
