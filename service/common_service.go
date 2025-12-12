package service

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func CalculateWorkingDays(tx *sqlx.Tx, start, end time.Time) (float64, error) {
	// 1️ Validate date range
	if end.Before(start) {
		return 0, fmt.Errorf("end date cannot be before start date")
	}

	// Normalize dates to midnight UTC to avoid timezone issues
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	// 2️ Fetch holidays within range
	var holidays []time.Time

	err := tx.Select(&holidays,
		`SELECT date FROM Tbl_Holiday 
         WHERE date BETWEEN $1 AND $2`,
		start, end)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch holidays: %v", err)
	}

	// Convert slice to a map for O(1) lookup
	holidayMap := make(map[string]bool)
	for _, h := range holidays {
		holidayMap[h.Format("2006-01-02")] = true
	}

	// 3️ Count working days
	workingDays := 0
	var workingDaysList []string // For debugging

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayStr := d.Format("2006-01-02")
		weekday := d.Weekday()

		// Skip Saturday and Sunday
		if weekday == time.Saturday || weekday == time.Sunday {
			fmt.Printf("DEBUG: Skipping weekend: %s (%s)\n", dayStr, weekday)
			continue
		}

		// Skip holidays
		if holidayMap[dayStr] {
			fmt.Printf("DEBUG: Skipping holiday: %s\n", dayStr)
			continue
		}

		workingDays++
		workingDaysList = append(workingDaysList, fmt.Sprintf("%s (%s)", dayStr, weekday))
	}

	fmt.Printf("DEBUG: Working days calculated: %d - Days: %v\n", workingDays, workingDaysList)
	return float64(workingDays), nil
}
