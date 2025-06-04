package main

import (
	"fmt"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/filedef"
)

func main() {
	f, err := os.Open("19313160934_ACTIVITY.fit")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := decoder.New(f)
	fitFile, err := dec.Decode()
	if err != nil {
		panic(err)
	}

	activity := filedef.NewActivity(fitFile.Messages...)

	fmt.Printf("Records count: %d\n", len(activity.Records))
	if len(activity.Records) > 0 {
		rec := activity.Records[0]
		fmt.Println("First record fields:")
		fmt.Printf("  Distance: %g m\n", rec.DistanceScaled())
		fmt.Printf("  Lat: %g degrees\n", rec.PositionLatDegrees())
		fmt.Printf("  Long: %g degrees\n", rec.PositionLongDegrees())
		fmt.Printf("  Speed: %g m/s\n", rec.SpeedScaled())
		fmt.Printf("  HeartRate: %d bpm\n", rec.HeartRate)
		fmt.Printf("  Cadence: %d rpm\n", rec.Cadence)
		fmt.Printf("  Timestamp: %v\n", rec.Timestamp)
	} else {
		fmt.Println("No records found.")
	}

	// --- Garmin UI-style Summary ---
	if len(activity.Sessions) > 0 {
		s := activity.Sessions[0]
		fmt.Printf("\n==== Garmin UI-Style Summary ====")

		// TIMING
		fmt.Printf("\n\nTiming\n------\n")
		fmt.Printf("Total Time: %.2f min\n", float64(s.TotalTimerTime)/60)
		fmt.Printf("Distance: %.2f km\n", float64(s.TotalDistance)/1000)

		// NUTRITION & HYDRATION
		fmt.Printf("\nNutrition & Hydration\n----------------------\n")
		fmt.Printf("Total Calories Burned: %d\n", s.TotalCalories)
		fmt.Printf("Active Calories: %d\n", s.TotalCalories)
		fmt.Printf("Resting Calories: (not in file, estimate from BMR × time)\n")
		fmt.Printf("Calories Consumed: (not in file)\n")
		fmt.Printf("Calories Net: (not in file, UI: Burned - Consumed)\n")
		fmt.Printf("Est. Sweat Loss: (not in file)\n")
		fmt.Printf("Fluid Consumed: (not in file)\n")
		fmt.Printf("Fluid Net: (not in file)\n")

		// TRAINING EFFECT
		fmt.Printf("\nTraining Effect\n---------------\n")
		fmt.Printf("Aerobic: %d\n", s.TotalTrainingEffect)
		fmt.Printf("Anaerobic: %d\n", s.TotalAnaerobicTrainingEffect)
		fmt.Printf("Exercise Load: %d\n", s.TrainingLoadPeak)
		fmt.Printf("Primary Benefit: %s\n", s.SportProfileName)

		// WORKOUT DETAILS
		fmt.Printf("\nWorkout Details\n---------------\n")
		fmt.Printf("Total Reps: (not in file)\n")
		fmt.Printf("Total Sets: (not in file)\n")
		fmt.Printf("Volume: (not in file)\n")

		// INTENSITY MINUTES
		fmt.Printf("\nIntensity Minutes\n-----------------\n")
		fmt.Printf("Moderate: (not in file, calculated by device/UI)\n")
		fmt.Printf("Vigorous: (not in file, calculated by device/UI)\n")
		fmt.Printf("Total: (not in file, calculated by device/UI)\n")

		// HEART RATE
		fmt.Printf("\nHeart Rate\n----------\n")
		fmt.Printf("Avg HR: %d bpm\n", s.AvgHeartRate)
		fmt.Printf("Max HR: %d bpm\n", s.MaxHeartRate)
	}

}
