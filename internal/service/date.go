package service

import (
	"fmt"
	"regexp"
	"time"
)

var monthYearRe = regexp.MustCompile(`^(0[1-9]|1[0-2])-\d{4}$`)

func ParseMonthYear(value string) (time.Time, error) {
	if !monthYearRe.MatchString(value) {
		return time.Time{}, fmt.Errorf("invalid month-year format: %q", value)
	}

	t, err := time.Parse("01-2006", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse month-year: %w", err)
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FormatMonthYear(t time.Time) string {
	return t.Format("01-2006")
}

func CountMonthsInIntersection(start time.Time, end *time.Time, from, to time.Time) int {
	start = normalizeMonth(start)
	from = normalizeMonth(from)
	to = normalizeMonth(to)

	var intersectionStart time.Time
	if start.After(from) {
		intersectionStart = start
	} else {
		intersectionStart = from
	}

	var intersectionEnd time.Time
	if end == nil {
		intersectionEnd = to
	} else {
		normalizedEnd := normalizeMonth(*end)
		if normalizedEnd.Before(to) {
			intersectionEnd = normalizedEnd
		} else {
			intersectionEnd = to
		}
	}

	if intersectionStart.After(intersectionEnd) {
		return 0
	}

	return (intersectionEnd.Year()-intersectionStart.Year())*12 +
		int(intersectionEnd.Month()-intersectionStart.Month()) + 1
}

func normalizeMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}