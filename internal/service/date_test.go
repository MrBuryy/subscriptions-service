package service

import (
	"testing"
	"time"
)

func TestParseMonthYear(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      time.Time
		wantError bool
	}{
		{
			name:  "valid may 2025",
			input: "05-2025",
			want:  time.Date(2025, time.May, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "valid december 2026",
			input: "12-2026",
			want:  time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
		{
			name:      "single digit month",
			input:     "5-2025",
			wantError: true,
		},
		{
			name:      "wrong order",
			input:     "2025-05",
			wantError: true,
		},
		{
			name:      "month 13",
			input:     "13-2025",
			wantError: true,
		},
		{
			name:      "month 00",
			input:     "00-2025",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMonthYear(tt.input)

			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !got.Equal(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatMonthYear(t *testing.T) {
	input := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.UTC)

	got := FormatMonthYear(input)

	if got != "05-2025" {
		t.Fatalf("got %q, want %q", got, "05-2025")
	}
}

func TestCountMonthsInIntersection(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   *time.Time
		from  time.Time
		to    time.Time
		want  int
	}{
		{
			name:  "full intersection",
			start: monthDate(2025, time.July),
			end:   timePtr(monthDate(2025, time.September)),
			from:  monthDate(2025, time.August),
			to:    monthDate(2025, time.September),
			want:  2,
		},
		{
			name:  "partial intersection on left",
			start: monthDate(2025, time.March),
			end:   timePtr(monthDate(2025, time.June)),
			from:  monthDate(2025, time.January),
			to:    monthDate(2025, time.April),
			want:  2,
		},
		{
			name:  "partial intersection on right",
			start: monthDate(2025, time.July),
			end:   timePtr(monthDate(2025, time.September)),
			from:  monthDate(2025, time.August),
			to:    monthDate(2025, time.December),
			want:  2,
		},
		{
			name:  "nil end",
			start: monthDate(2025, time.July),
			end:   nil,
			from:  monthDate(2025, time.August),
			to:    monthDate(2025, time.October),
			want:  3,
		},
		{
			name:  "no intersection",
			start: monthDate(2025, time.July),
			end:   timePtr(monthDate(2025, time.September)),
			from:  monthDate(2025, time.October),
			to:    monthDate(2025, time.December),
			want:  0,
		},
		{
			name:  "subscription starts before period and no end",
			start: monthDate(2025, time.March),
			end:   nil,
			from:  monthDate(2025, time.January),
			to:    monthDate(2025, time.April),
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountMonthsInIntersection(tt.start, tt.end, tt.from, tt.to)
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func monthDate(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func timePtr(t time.Time) *time.Time {
	return &t
}