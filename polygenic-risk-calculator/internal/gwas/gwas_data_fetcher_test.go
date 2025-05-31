package gwas

import (
	"testing"
	"phite.io/polygenic-risk-calculator/internal/model"
	"reflect"

	"phite.io/polygenic-risk-calculator/internal/logging"
)


func TestFetchAndAnnotateGWAS_BasicAnnotation(t *testing.T) {
	logging.SetSilentLoggingForTest()
	input := GWASDataFetcherInput{
		ValidatedSNPs: []model.ValidatedSNP{
			{RSID: "rs1", Genotype: "AA", FoundInGWAS: true},
			{RSID: "rs2", Genotype: "AG", FoundInGWAS: true},
		},
		AssociationsClean: []model.GWASSNPRecord{
			{RSID: "rs1", RiskAllele: "A", Beta: 0.2, Trait: "height"},
			{RSID: "rs2", RiskAllele: "G", Beta: -0.1, Trait: "height"},
		},
	}
	want := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.2, Dosage: 2, Trait: "height"},
		{RSID: "rs2", Genotype: "AG", RiskAllele: "G", Beta: -0.1, Dosage: 1, Trait: "height"},
	}
	out := FetchAndAnnotateGWAS(input)
	if !reflect.DeepEqual(out.AnnotatedSNPs, want) {
		t.Errorf("Basic annotation failed. Got %+v, want %+v", out.AnnotatedSNPs, want)
	}
}

func TestFetchAndAnnotateGWAS_MissingGWAS(t *testing.T) {
	logging.SetSilentLoggingForTest()
	input := GWASDataFetcherInput{
		ValidatedSNPs: []model.ValidatedSNP{
			{RSID: "rs3", Genotype: "TT", FoundInGWAS: false},
		},
		AssociationsClean: []model.GWASSNPRecord{},
	}
	out := FetchAndAnnotateGWAS(input)
	if len(out.AnnotatedSNPs) != 0 {
		t.Errorf("Expected no annotations for missing GWAS, got %+v", out.AnnotatedSNPs)
	}
}

func TestFetchAndAnnotateGWAS_MultipleAssociations(t *testing.T) {
	logging.SetSilentLoggingForTest()
	input := GWASDataFetcherInput{
		ValidatedSNPs: []model.ValidatedSNP{
			{RSID: "rs4", Genotype: "CC", FoundInGWAS: true},
		},
		AssociationsClean: []model.GWASSNPRecord{
			{RSID: "rs4", RiskAllele: "C", Beta: 0.3, Trait: "BMI"},
			{RSID: "rs4", RiskAllele: "C", Beta: 0.2, Trait: "height"},
		},
	}
	out := FetchAndAnnotateGWAS(input)
	if len(out.AnnotatedSNPs) != 2 {
		t.Errorf("Expected two annotations for multiple associations, got %+v", out.AnnotatedSNPs)
	}
}

func TestFetchAndAnnotateGWAS_AmbiguousGenotype(t *testing.T) {
	logging.SetSilentLoggingForTest()
	input := GWASDataFetcherInput{
		ValidatedSNPs: []model.ValidatedSNP{
			{RSID: "rs5", Genotype: "NN", FoundInGWAS: true},
		},
		AssociationsClean: []model.GWASSNPRecord{
			{RSID: "rs5", RiskAllele: "N", Beta: 0.1, Trait: "height"},
		},
	}
	out := FetchAndAnnotateGWAS(input)
	if len(out.AnnotatedSNPs) != 1 || out.AnnotatedSNPs[0].Dosage != 0 {
		t.Errorf("Ambiguous genotype should yield dosage 0. Got %+v", out.AnnotatedSNPs)
	}
}
