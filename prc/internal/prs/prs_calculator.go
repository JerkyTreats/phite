package prs

import (
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/invariance"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
)

// Canonical structs from data_model.md

type SNPContribution struct {
	Rsid         string
	Dosage       int
	Beta         float64
	Contribution float64
}

type PRSResult struct {
	PRSScore float64
	Details  []SNPContribution
}

// PRSCalculationError represents errors that occur during PRS calculation with invariance violations
type PRSCalculationError struct {
	Message string
	SNP     *model.AnnotatedSNP
	Phase   string // "pre-condition", "calculation", "post-condition"
	Cause   error
}

func (e *PRSCalculationError) Error() string {
	if e.SNP != nil {
		return fmt.Sprintf("PRS calculation error in %s phase for SNP %s: %s (cause: %v)",
			e.Phase, e.SNP.RSID, e.Message, e.Cause)
	}
	return fmt.Sprintf("PRS calculation error in %s phase: %s (cause: %v)",
		e.Phase, e.Message, e.Cause)
}

func (e *PRSCalculationError) Unwrap() error {
	return e.Cause
}

// CalculatePRS computes the polygenic risk score for a set of SNPs with optional invariance validation.
func CalculatePRS(snps []model.AnnotatedSNP) (PRSResult, error) {
	logging.Info("Starting PRS calculation for %d SNPs", len(snps))

	// Pre-condition validation - self-contained check
	if err := ValidateInputSNPs(snps); err != nil {
		return PRSResult{}, &PRSCalculationError{
			Message: "Input validation failed",
			SNP:     nil,
			Phase:   "pre-condition",
			Cause:   err,
		}
	}

	// Core PRS calculation logic
	var total float64
	contributions := make([]SNPContribution, 0, len(snps))

	for _, snp := range snps {
		contribution := float64(snp.Dosage) * snp.Beta

		// Optional runtime validation for individual contributions
		if err := ValidateVariantContribution(snp.RSID, snp.Dosage, snp.Beta, contribution); err != nil {
			return PRSResult{}, &PRSCalculationError{
				Message: "Variant contribution validation failed",
				SNP:     &snp,
				Phase:   "calculation",
				Cause:   err,
			}
		}

		contributions = append(contributions, SNPContribution{
			Rsid:         snp.RSID,
			Dosage:       snp.Dosage,
			Beta:         snp.Beta,
			Contribution: contribution,
		})
		total += contribution
	}

	result := PRSResult{
		PRSScore: total,
		Details:  contributions,
	}

	// Post-condition validation - self-contained check
	if err := ValidatePRSResult(snps, result); err != nil {
		return PRSResult{}, &PRSCalculationError{
			Message: "Result validation failed",
			SNP:     nil,
			Phase:   "post-condition",
			Cause:   err,
		}
	}

	logging.Info("PRS calculation complete: score=%v, SNPs=%d", result.PRSScore, len(result.Details))
	return result, nil
}

// CalculatePRSWithBounds computes PRS with theoretical bounds validation
func CalculatePRSWithBounds(snps []model.AnnotatedSNP, theoreticalMin, theoreticalMax float64) (PRSResult, error) {
	result, err := CalculatePRS(snps)
	if err != nil {
		return result, err
	}

	// Additional bounds validation - self-contained check
	if err := ValidatePRSBounds(result.PRSScore, theoreticalMin, theoreticalMax); err != nil {
		return PRSResult{}, &PRSCalculationError{
			Message: "PRS bounds validation failed",
			SNP:     nil,
			Phase:   "post-condition",
			Cause:   err,
		}
	}

	logging.Info("PRS bounds validation passed: %v ∈ [%v, %v]", result.PRSScore, theoreticalMin, theoreticalMax)
	return result, nil
}

// IsValidationEnabled returns whether invariance validation is currently enabled
func IsValidationEnabled() bool {
	return invariance.IsValidationEnabled()
}

// IsStrictModeEnabled returns whether strict validation mode is currently enabled
func IsStrictModeEnabled() bool {
	return invariance.IsStrictModeEnabled()
}
