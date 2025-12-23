package main

import (
	"testing"
	"time"
)

func TestFormatFilename(t *testing.T) {
	tests := []struct {
		name             string
		pattern          string
		originalFilename string
		emailDate        string
		want             string
	}{
		{
			name:             "replace original placeholder",
			pattern:          "prefix_{original}_suffix",
			originalFilename: "test.pdf",
			emailDate:        "2024-01-01",
			want:             "prefix_test.pdf_suffix",
		},
		{
			name:             "replace date placeholder",
			pattern:          "file_{date}.pdf",
			originalFilename: "test.pdf",
			emailDate:        "2024-01-01_12-00-00",
			want:             "file_2024-01-01_12-00-00.pdf",
		},
		{
			name:             "replace both placeholders",
			pattern:          "{date}_{original}",
			originalFilename: "document.pdf",
			emailDate:        "2024-01-01_12-00-00",
			want:             "2024-01-01_12-00-00_document.pdf",
		},
		{
			name:             "no placeholders",
			pattern:          "static_filename.pdf",
			originalFilename: "test.pdf",
			emailDate:        "2024-01-01",
			want:             "static_filename.pdf",
		},
		{
			name:             "empty pattern",
			pattern:          "",
			originalFilename: "test.pdf",
			emailDate:        "2024-01-01",
			want:             "",
		},
		{
			name:             "multiple occurrences",
			pattern:          "{original}_{date}_{original}",
			originalFilename: "file.pdf",
			emailDate:        "2024-01-01",
			want:             "file.pdf_2024-01-01_file.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFilename(tt.pattern, tt.originalFilename, tt.emailDate)
			if got != tt.want {
				t.Errorf("formatFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEmailDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		want    string
	}{
		{
			name:    "RFC1123Z format",
			dateStr: "Mon, 02 Jan 2006 15:04:05 -0700",
			want:    "2006-01-02_15-04-05",
		},
		{
			name:    "RFC1123 format",
			dateStr: "Mon, 02 Jan 2006 15:04:05 MST",
			want:    "2006-01-02_15-04-05",
		},
		{
			name:    "single digit day",
			dateStr: "Mon, 2 Jan 2006 15:04:05 -0700",
			want:    "2006-01-02_15-04-05",
		},
		{
			name:    "no weekday",
			dateStr: "2 Jan 2006 15:04:05 -0700",
			want:    "2006-01-02_15-04-05",
		},
		{
			name:    "invalid date",
			dateStr: "invalid date string",
			want:    "unknown",
		},
		{
			name:    "empty string",
			dateStr: "",
			want:    "unknown",
		},
		{
			name:    "realistic email date",
			dateStr: "Wed, 15 Jan 2024 10:30:45 +0530",
			want:    "2024-01-15_10-30-45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEmailDate(tt.dateStr)
			if got != tt.want {
				t.Errorf("parseEmailDate(%q) = %v, want %v", tt.dateStr, got, tt.want)
			}
		})
	}
}

func TestParseEmailDate_RealWorldExamples(t *testing.T) {
	// Test with various real-world email date formats
	realWorldDates := []struct {
		dateStr string
		want    string
	}{
		{
			dateStr: "Mon, 1 Jan 2024 12:00:00 +0000",
			want:    "2024-01-01_12-00-00",
		},
		{
			dateStr: "Tue, 31 Dec 2023 23:59:59 -0800",
			want:    "2023-12-31_23-59-59",
		},
		{
			dateStr: "Fri, 15 Mar 2024 08:15:30 +0530",
			want:    "2024-03-15_08-15-30",
		},
	}

	for _, tt := range realWorldDates {
		t.Run(tt.dateStr, func(t *testing.T) {
			got := parseEmailDate(tt.dateStr)
			if got != tt.want {
				t.Errorf("parseEmailDate(%q) = %v, want %v", tt.dateStr, got, tt.want)
			}
		})
	}
}

func TestParseEmailDate_ConsistentFormat(t *testing.T) {
	// Test that the output format is consistent
	dateStr := "Mon, 02 Jan 2006 15:04:05 -0700"
	result := parseEmailDate(dateStr)

	// Verify format: YYYY-MM-DD_HH-MM-SS
	if len(result) != 19 {
		t.Errorf("parseEmailDate() result length = %d, want 19 (format: YYYY-MM-DD_HH-MM-SS)", len(result))
	}

	// Try to parse the result back to verify it's a valid date format
	_, err := time.Parse("2006-01-02_15-04-05", result)
	if err != nil {
		t.Errorf("parseEmailDate() result %q is not in expected format: %v", result, err)
	}
}

