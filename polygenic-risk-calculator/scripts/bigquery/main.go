package main

import (
	"fmt"

	"phite.io/polygenic-risk-calculator/scripts/bigquery/gnomad"
	"phite.io/polygenic-risk-calculator/scripts/bigquery/jerkytreats"
)

func main() {
	fmt.Println("Verifying gnomAD schema...")
	gnomad.VerifySchema()

	fmt.Println("\nVerifying Jerkytreats schema...")
	jerkytreats.VerifySchema()
}
